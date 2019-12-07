package main

type args struct {
	data []string
}

func (args *args) push(data ...string) {
	args.data = append(args.data, data...)
}
