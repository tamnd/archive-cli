package ia

import (
	"context"
	"encoding/json"
	"net/url"
)

// Task is one entry from an item's catalog/derive history. The complete raw
// record is retained so no task field is dropped in json/jsonl output.
type Task struct {
	TaskID    int64  `json:"task_id"`
	Server    string `json:"server"`
	Cmd       string `json:"cmd"`
	Status    string `json:"status"`
	Args      any    `json:"args"`
	Category  string `json:"category"`
	Priority  int    `json:"priority"`
	Submitter string `json:"submitter"`
	DateSub   string `json:"submittime"`
	Finished  int64  `json:"finished"`
	Notes     string `json:"notes"`

	Raw json.RawMessage `json:"-"`
}

// UnmarshalJSON fills the typed fields and retains the complete raw record.
func (t *Task) UnmarshalJSON(b []byte) error {
	type alias Task // sheds UnmarshalJSON to avoid recursion
	var a alias
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	*t = Task(a)
	t.Raw = append(t.Raw[:0], b...)
	return nil
}

// Fields decodes the complete task record so every field survives into output.
func (t Task) Fields() map[string]any {
	out := map[string]any{}
	if len(t.Raw) > 0 {
		_ = json.Unmarshal(t.Raw, &out)
	}
	return out
}

// GetTasks fetches the catalog task history for an item. The tasks endpoint
// requires credentials for items you do not own; the client attaches them when
// present.
func GetTasks(ctx context.Context, h *HTTPClient, identifier string) ([]Task, error) {
	v := url.Values{"identifier": {identifier}}
	u := TasksURL + "?" + v.Encode()
	b, err := h.FetchBytes(ctx, u)
	if err != nil {
		return nil, err
	}
	var r struct {
		Value struct {
			History []Task `json:"history"`
			Catalog []Task `json:"catalog"`
		} `json:"value"`
	}
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, err
	}
	out := append([]Task{}, r.Value.Catalog...)
	out = append(out, r.Value.History...)
	return out, nil
}
