package kit

// InterceptApply TODO
func InterceptApply(o KitOptions) {
	stopper := make(chan struct{})
	o.Stopper = stopper
	// for _, obj := range o.Objects {
	// 	fmt.Printf("%+v\n\n", obj.Object)
	// 	obj.Get()
	// 	fmt.Printf("%+v\n\n", obj.Object)
	// }

	UI(o)
}
