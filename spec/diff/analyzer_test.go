package diff

import (
	_ "embed"
	"reflect"
	"testing"
)

var (
	//go:embed testdata/swagger_single_op.json
	swaggerSingleOp []byte

	//go:embed testdata/swagger_two_op.json
	swaggerTwoOp []byte
)

func TestAnalyze(t *testing.T) {
	type args struct {
		fromSpecJSON []byte
		toSpecJSON   []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *Report
		wantErr bool
	}{
		{
			name: "adding an operation reports it without error",
			args: args{
				fromSpecJSON: swaggerSingleOp,
				toSpecJSON:   swaggerTwoOp,
			},
			want: func() *Report {
				r := NewReport()
				// TODO: set expectations on the report. Right now we just
				//  want to see the output, so this will print the actual
				//  report when the test fails.
				return r
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Analyze(tt.args.fromSpecJSON, tt.args.toSpecJSON)
			if (err != nil) != tt.wantErr {
				t.Errorf("Analyze() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Analyze() got = %v, want %v", got, tt.want)
			}
		})
	}
}
