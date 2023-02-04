package utils

func ParamsToURL(params map[string]string) string {
	var url string
	for key, value := range params {
		url += "&" + key + "=" + value
	}
	return url
}
