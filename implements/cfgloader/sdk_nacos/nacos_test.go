package sdk_nacos

import (
	"context"
	"fmt"
	"testing"
)

var c = NewClient(NacosConfig{
	AuthConfig: AuthConfig{
		Username: "nacos",
		Password: "OoGZMPoOz8ydAAGb837Hl6nCwugNlgvhXTkQtW4b",
	},
	Endpoint: "http://nacos-dev.39on.com",
})
var testID = "9523f13f-1e4a-42e3-8076-00bf316eff4e"

func Test_client_UpdateConfig(t *testing.T) {
	type args struct {
		ctx  context.Context
		data ConfigData
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "",
			args: args{
				ctx: context.Background(),
				data: ConfigData{
					TenantID: testID,
					DataID:   "test",
					GroupID:  "test2",
					Content:  `{"asasg":"{$test.test3}","fff":{"$ref":"test"}}`,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := c.UpdateConfig(tt.args.ctx, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("UpdateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_client_GetConfig(t *testing.T) {
	type args struct {
		ctx  context.Context
		data ConfigData
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "",
			args: args{
				ctx: context.Background(),
				data: ConfigData{
					TenantID: testID,
					DataID:   "test",
					GroupID:  "test2",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ret, err := c.GetConfig(tt.args.ctx, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			t.Logf("%s", ret)
		})
	}
}

func Test_client_ListenConfig(t *testing.T) {
	type args struct {
		ctx  context.Context
		data ConfigData
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "",
			args: args{
				ctx: context.Background(),
				data: ConfigData{
					DataID:   "http-service-demo",
					GroupID:  "sy-go",
					TenantID: testID,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ret, err := c.GetConfig(tt.args.ctx, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			t.Logf("%s", ret)
			tt.args.data.Content = string(ret)
			res, ok, err := c.ListenConfig(tt.args.ctx, ListenConfigData{
				TenantID: testID,
				Configs:  []ConfigData{tt.args.data, tt.args.data},
			})
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("%s %v", res, ok)
		})
	}
}

func Test_client_GetRefConfig(t *testing.T) {
	type args struct {
		ctx  context.Context
		data ConfigData
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "",
			args: args{
				ctx: context.Background(),
				data: ConfigData{
					TenantID: testID,
					DataID:   "test",
					GroupID:  "test2",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ret, err := c.GetRefConfig(tt.args.ctx, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			t.Logf("%s", ret)
		})
	}
}

func Test_client_DelConfig(t *testing.T) {
	type args struct {
		ctx  context.Context
		data ConfigData
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "",
			args: args{
				ctx: context.Background(),
				data: ConfigData{
					TenantID: testID,
					DataID:   "test",
					GroupID:  "test2",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := c.DelConfig(tt.args.ctx, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("UpdateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_client_GetAccessToken(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "获取accessToken",
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if ak := c.GetAccessToken(tt.args.ctx); ak != "" {
				fmt.Println(ak)
			}
		})
	}
}
