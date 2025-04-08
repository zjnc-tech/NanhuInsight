package api

const (
	// Success 成功
	Success int = 200
	// ParamError 参数异常
	ParamError int = 400
	// InternalError 服务器内部异常
	InternalError int = 500
)

type CommonResponse struct {
	Code      int         `json:"code,omitempty"`
	Message   string      `json:"message,omitempty"`
	RequestId interface{} `json:"requestId,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Success   bool        `json:"success,omitempty"`
	Failed    bool        `json:"failed,omitempty"`
}

func SuccessResponse(data interface{}) CommonResponse {
	return CommonResponse{
		Code:    Success,
		Message: "success",
		Data:    data,
	}
}

func SuccessResponseWithMsg(data interface{}, msg string) CommonResponse {
	return CommonResponse{
		Code:    Success,
		Message: msg,
		Data:    data,
	}
}

func ParamErrResponse(msg string) CommonResponse {
	return CommonResponse{
		Code:    ParamError,
		Message: msg,
	}
}

func InternalErrResponse(msg string) CommonResponse {
	return CommonResponse{
		Code:    InternalError,
		Message: msg,
	}
}
