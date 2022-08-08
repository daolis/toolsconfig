package toolsconfig

import (
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_defaultConfigFileName(t *testing.T) {
	type args struct {
		dir  string
		file string
	}
	userHomedir, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "relative path",
			args: args{
				dir:  ".",
				file: "testfile",
			},
			want:    "testfile",
			wantErr: false,
		},
		{
			name: "absolute path",
			args: args{
				dir:  "/tmp",
				file: "testfile",
			},
			want:    "/tmp/testfile",
			wantErr: false,
		},
		{
			name: "homedir path",
			args: args{
				dir:  ".toolsconfig",
				file: "testfile",
			},
			want:    userHomedir + "/.toolsconfig/testfile",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := defaultConfigFileName(tt.args.dir, tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("defaultConfigFileName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.NotNil(t, got)
			if !reflect.DeepEqual(*got, tt.want) {
				t.Errorf("defaultConfigFileName() got = %v, want %v", got, tt.want)
			}
		})
	}
}
