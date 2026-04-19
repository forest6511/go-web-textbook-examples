package apperror

// Problem は RFC 7807 Problem Details の JSON 表現
type Problem struct {
	Type     string       `json:"type"`               // 問題タイプ URI
	Title    string       `json:"title"`              // 短い要約
	Status   int          `json:"status"`             // HTTP ステータス
	Detail   string       `json:"detail,omitempty"`   // 詳細説明
	Instance string       `json:"instance,omitempty"` // 発生事例の識別子
	Code     string       `json:"code"`               // 拡張: 安定コード
	TraceID  string       `json:"trace_id,omitempty"` // 拡張: トレース ID
	Errors   []FieldIssue `json:"errors,omitempty"`   // 拡張: バリデーション詳細
}

// ToProblem は *AppError を Problem に変換する
func ToProblem(e *AppError, traceID string) Problem {
	return Problem{
		Type:    "about:blank",
		Title:   e.Code,
		Status:  e.HTTPCode,
		Detail:  e.Message,
		Code:    e.Code,
		TraceID: traceID,
		Errors:  e.Details,
	}
}
