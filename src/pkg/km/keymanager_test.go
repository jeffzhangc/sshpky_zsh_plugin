package km_test

import (
	"sshpky/pkg/km"
	"strings"
	"testing"
)

func TestGetPassword(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		username string
		host     string
		want     string
		wantErr  bool
	}{
		// TODO: Add test cases.
		{
			name:     "t1",
			username: "zhanghl",
			host:     "aicc_dev_test",
			want:     "abc123123",
			wantErr:  false,
		},
		{
			name:     "test pwd donot exist",
			username: "zhanghl",
			host:     "aicc_dev_test_exist",
			want:     "",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := km.GetPassword(tt.username, tt.host)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetPassword() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetPassword() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			// if tt.want == got {
			if !strings.EqualFold(got, tt.want) {
				t.Errorf("GetPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSavePassword(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		username string
		host     string
		password string
		wantErr  bool
	}{
		// TODO: Add test cases.
		{
			name:     "test add",
			username: "zhanghl",
			host:     "aicc_dev_test",
			password: "abc123123",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := km.SavePassword(tt.username, tt.host, tt.password)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("SavePassword() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("SavePassword() succeeded unexpectedly")
			}
		})
	}
}

func TestGenerateOTP(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		secret  string
		want    string
		wantErr bool
	}{
		// TODO: Add test cases
		{
			name:    "t1",
			secret:  "PCZEC3XDHKK42Y2Y",
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := km.GenerateOTP(tt.secret)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GenerateOTP() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GenerateOTP() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("GenerateOTP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetMFASecret(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		user    string
		host    string
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "t1",
			user: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := km.GetMFASecret(tt.user, tt.host)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetMFASecret() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetMFASecret() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("GetMFASecret() = %v, want %v", got, tt.want)
			}
		})
	}
}
