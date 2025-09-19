package sdk_acm

import (
	"context"
	"testing"
	"time"
)

var cli = NewClient("http://scm-dev.39on.com")

func Test_client_GetConfigBreak(t *testing.T) {
	go func() {
	}()
	ctx := context.Background()
	for {
		for i := 0; i < 2; i++ {
			go func(ctx context.Context) {
				ctx, cf := context.WithTimeout(ctx, time.Millisecond*200)
				defer cf()
				_, err := cli.GetConfig(ctx, ConfigData{
					TenantID:  "xxx",
					DataID:    "xx",
					GroupID:   "xx",
					AppID:     "xxx",
					Namespace: "xx",
					Content:   "xx",
				})
				if err != nil {
					t.Log(err)
				} else {
					t.Log("success")
				}
			}(ctx)
		}
		time.Sleep(time.Second * 1)
	}
}

func Test_client_GetConfig(t *testing.T) {
	type args struct {
		data ConfigData
	}
	tests := []struct {
		name        string
		args        args
		wantContent []byte
		wantErr     bool
	}{
		{
			args: args{
				data: ConfigData{
					TenantID: "9523f13f-1e4a-42e3-8076-00bf316eff4e",
					DataID:   "application-config-manager",
					GroupID:  "go-devops",
				},
			},
		},
		{
			args: args{
				data: ConfigData{
					TenantID: "9523f13f-1e4a-42e3-8076-00bf316eff4e",
					//DataID:   "application-config-manager",
					GroupID: "go-devops",
				},
			},
		},
		{
			args: args{
				data: ConfigData{
					TenantID: "9523f13f-1e4a-42e3-8076-00bf316eff4e",
					DataID:   "application-config-manager",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotContent, err := cli.GetConfig(context.Background(), tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("%s", gotContent)
		})
	}
}

func Test_client_UpdateConfig(t *testing.T) {
	type args struct {
		data ConfigData
	}
	tests := []struct {
		name        string
		args        args
		wantContent []byte
		wantErr     bool
	}{
		{
			args: args{
				data: ConfigData{
					TenantID: "9523f13f-1e4a-42e3-8076-00bf316eff4e",
					DataID:   "test",
					GroupID:  "test2",
					Content:  `{"test":"tttt"}`,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cli.UpdateConfig(context.Background(), tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_client_UpdateConfigSet(t *testing.T) {
	type args struct {
		data ConfigSet
	}
	tests := []struct {
		name        string
		args        args
		wantContent []byte
		wantErr     bool
	}{
		{
			args: args{
				data: ConfigSet{
					TenantID: "9523f13f-1e4a-42e3-8076-00bf316eff4e",
					DataIDs:  []string{"test"},
					GroupID:  "test2",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cli.UpdateConfigSet(context.Background(), tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
