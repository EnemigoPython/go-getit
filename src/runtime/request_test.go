package runtime

import (
	"testing"
)

func TestActionString(t *testing.T) {
	action := Store
	expected := "Store"
	if action.String() != expected {
		t.Errorf("Expected %s, got %s", expected, action.String())
	}
}

func TestRequestConstruction(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantErr    bool
		wantAction Action
	}{
		{
			name:       "valid store command",
			args:       []string{"store", "key1", "value1"},
			wantErr:    false,
			wantAction: Store,
		},
		{
			name:    "invalid empty command",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := ConstructRequest(tt.args, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConstructRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && req.GetAction() != tt.wantAction {
				t.Errorf("ConstructRequest() action = %v, want %v", req.GetAction(), tt.wantAction)
			}
		})
	}
}
