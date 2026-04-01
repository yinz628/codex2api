package admin

import (
	"io"
	"net/http"
	"strings"
)

const (
	proxyQualityStatusWarn = "warn"
	proxyQualityStatusFail = "fail"
)

type proxyQualityResult struct {
	Status     string `json:"quality_status"`
	StatusCode int    `json:"quality_status_code"`
}

type proxyTestResponse struct {
	Success           bool   `json:"success"`
	IP                string `json:"ip,omitempty"`
	Country           string `json:"country,omitempty"`
	Region            string `json:"region,omitempty"`
	City              string `json:"city,omitempty"`
	ISP               string `json:"isp,omitempty"`
	LatencyMs         int    `json:"latency_ms,omitempty"`
	Location          string `json:"location,omitempty"`
	TestCountry       string `json:"test_country,omitempty"`
	QualityStatus     string `json:"quality_status,omitempty"`
	QualityStatusCode int    `json:"quality_status_code,omitempty"`
	Error             string `json:"error,omitempty"`
}

func buildProxyLocation(country, region, city string) string {
	parts := make([]string, 0, 3)
	for _, part := range []string{country, region, city} {
		part = strings.TrimSpace(part)
		if part != "" {
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, " / ")
}

func normalizeProxyCountry(country, location string) string {
	country = strings.TrimSpace(country)
	if country != "" {
		return country
	}
	return extractCountryFromLocation(location)
}

func extractCountryFromLocation(location string) string {
	location = strings.TrimSpace(location)
	if location == "" {
		return ""
	}
	for _, sep := range []string{" / ", "·", "•", "|", ",", "，", ":", "："} {
		if idx := strings.Index(location, sep); idx > 0 {
			return strings.TrimSpace(location[:idx])
		}
	}
	return location
}

func runProxyQualityCheck(client *http.Client) proxyQualityResult {
	if client == nil {
		return proxyQualityResult{Status: proxyQualityStatusFail}
	}

	req, err := http.NewRequest(http.MethodGet, "https://api.openai.com/v1/models", nil)
	if err != nil {
		return proxyQualityResult{Status: proxyQualityStatusFail}
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return proxyQualityResult{Status: proxyQualityStatusFail}
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))

	if resp.StatusCode == http.StatusUnauthorized {
		return proxyQualityResult{
			Status:     proxyQualityStatusWarn,
			StatusCode: resp.StatusCode,
		}
	}

	return proxyQualityResult{
		Status:     proxyQualityStatusFail,
		StatusCode: resp.StatusCode,
	}
}
