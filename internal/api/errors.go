package api

import "fmt"

type HTTPError struct {
	StatusCode int
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("http error: %d", e.StatusCode)
}

func HumanHTTPError(status int) string {
	switch status {
	case 400:
		return "400 Bad Request: Some required fields are missing or have invalid values."
	case 401:
		return "401 Unauthorized: The required 'apiKey' field value is missing or invalid."
	case 403:
		return "403 Forbidden: Access restricted. Check credits balance or enter the correct API key."
	case 408:
		return "408 Request Timeout: The request has timed out. Try your call again later."
	case 410:
		return "410 Gone: The requested API version is no longer available. Update to the latest version of the API."
	case 422:
		return "422 Unprocessable Entity: Input correct request parameters or search term."
	case 429:
		return "429 Too Many Requests: Too Many Requests. Try your call again later."
	default:
		if status >= 500 && status <= 599 {
			return "5XX Internal Server Error: Internal server error, please contact the provider support team."
		}
		return fmt.Sprintf("Unexpected HTTP status: %d", status)
	}
}
