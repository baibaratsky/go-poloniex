package poloniex

import "testing"

func Test_convertibleBool_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    bool
	}{
		{"TrueNumberValue", args{[]byte(`"1"`)}, false, true},
		{"TrueValue", args{[]byte(`"true"`)}, false, true},
		{"FalseNumberValue", args{[]byte(`"0"`)}, false, false},
		{"FalseValue", args{[]byte(`"false"`)}, false, false},
		{"ErrorValue", args{[]byte(`"//"`)}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cbool := new(convertibleBool)
			err := cbool.UnmarshalJSON(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertibleBool.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			if bool(*cbool) != tt.want {
				t.Errorf("convertibleBool.UnmarshalJSON() value = %v, want %v", bool(*cbool), tt.want)
			}
		})
	}
}
