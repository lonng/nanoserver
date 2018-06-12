package protocol

type TestRequest struct {
	IntField    int    `json:"int_field"`
	StringField string `json:"string_field"`
}

type TestMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
