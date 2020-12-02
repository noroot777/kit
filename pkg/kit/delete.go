package kit

// InterceptDelete TODO
func InterceptDelete(o *Options) {
	o.stopper = make(chan struct{})
	defer func() { o.stopper <- struct{}{} }()
	// o.PrintFunc = printProcess

	DrawUI(o)
}

// func printProcess(objets []*resource.Info, txtView *ui.TextView) {
// 	for _, obj := range objets {
// 		var txt []string
// 		txt = append(txt, fmt.Sprint("%s \"%s\" %s\n", kindString, info.Name, operation))
// 		txtView.AddText(txt)
// 	}

// }

// TODO change the output and errorput of current cmd
