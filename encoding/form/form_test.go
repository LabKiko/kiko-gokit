package form

import (
	"reflect"
	"testing"

	"github.com/LabKiko/kiko-gokit/encoding"
)

type LoginRequest struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type TestModel struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

const contentType = "x-www-form-urlencoded"

func TestFormCodecMarshal(t *testing.T) {
	req := &LoginRequest{
		Username: "gokit",
		Password: "gokit_pwd",
	}
	content, err := encoding.GetCodec(contentType).Marshal(req)
	if err != nil {
		t.Errorf("marshal error: %v", err)
	}
	if !reflect.DeepEqual([]byte("password=gokit_pwd&username=gokit"), content) {
		t.Errorf("expect %v, got %v", []byte("password=gokit_pwd&username=gokit"), content)
	}

	req = &LoginRequest{
		Username: "gokit",
		Password: "",
	}
	content, err = encoding.GetCodec(contentType).Marshal(req)
	if err != nil {
		t.Errorf("expect %v, got %v", nil, err)
	}
	if !reflect.DeepEqual([]byte("username=gokit"), content) {
		t.Errorf("expect %v, got %v", []byte("username=gokit"), content)
	}

	m := &TestModel{
		ID:   1,
		Name: "gokit",
	}
	content, err = encoding.GetCodec(contentType).Marshal(m)
	t.Log(string(content))
	if err != nil {
		t.Errorf("expect %v, got %v", nil, err)
	}
	if !reflect.DeepEqual([]byte("id=1&name=gokit"), content) {
		t.Errorf("expect %v, got %v", []byte("id=1&name=gokit"), content)
	}
}

func TestFormCodecUnmarshal(t *testing.T) {
	req := &LoginRequest{
		Username: "gokit",
		Password: "gokit_pwd",
	}
	content, err := encoding.GetCodec(contentType).Marshal(req)
	if err != nil {
		t.Errorf("expect %v, got %v", nil, err)
	}

	bindReq := new(LoginRequest)
	err = encoding.GetCodec(contentType).Unmarshal(content, bindReq)
	if err != nil {
		t.Errorf("expect %v, got %v", nil, err)
	}
	if !reflect.DeepEqual("gokit", bindReq.Username) {
		t.Errorf("expect %v, got %v", "gokit", bindReq.Username)
	}
	if !reflect.DeepEqual("gokit_pwd", bindReq.Password) {
		t.Errorf("expect %v, got %v", "gokit_pwd", bindReq.Password)
	}
}
