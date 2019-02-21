package controllers

// Define response json data format, In response to a data format defined json
type ResponseInfo struct {
	Code    string
	Message string
	Data    interface{}
}
