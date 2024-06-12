package internal

func ExamplePrintBuildInfo() {
	type args struct {
		version string
		date    string
		commit  string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "1",
			args: args{
				version: "1",
				date:    "01/01/01",
				commit:  "34534",
			},
		},
		{
			name: "2",
			args: args{
				version: "",
				date:    "",
				commit:  "",
			},
		},
	}
	for _, tt := range tests {
		PrintBuildInfo(tt.args.version, tt.args.date, tt.args.commit)
	}

	// Output:
	// Build version: 1
	// Build date: 01/01/01
	// Build commit: 34534
	// Build version: N/A
	// Build date: N/A
	// Build commit: N/A
}
