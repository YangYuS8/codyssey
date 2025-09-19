package errcode

// Code 定义统一错误码常量
const (
    CodeUnauthorized = "UNAUTHORIZED"
    CodeForbidden    = "FORBIDDEN"
    CodeInvalidID    = "INVALID_ID"
    CodeNotFound     = "NOT_FOUND"
    CodeSubmissionNotFound = "SUBMISSION_NOT_FOUND"
    CodeEnqueueFailed = "ENQUEUE_FAILED"
    CodeListFailed    = "LIST_FAILED"
    // JudgeRun 扩展
    CodeJudgeRunNotFound   = "JUDGE_RUN_NOT_FOUND"
    CodeInvalidStatus      = "INVALID_STATUS"
    CodeInvalidTransition  = "INVALID_TRANSITION"
)

// Msg 提供默认错误消息，可在 handler 中覆盖
var Msg = map[string]string{
    CodeUnauthorized:       "login required",
    CodeForbidden:          "forbidden",
    CodeInvalidID:          "invalid identifier",
    CodeNotFound:           "resource not found",
    CodeSubmissionNotFound: "submission not found",
    CodeEnqueueFailed:      "enqueue failed",
    CodeListFailed:         "list failed",
    CodeJudgeRunNotFound:   "judge run not found",
    CodeInvalidStatus:      "invalid status value",
    CodeInvalidTransition:  "invalid status transition",
}

func Text(code string) string {
    if s, ok := Msg[code]; ok { return s }
    return code
}
