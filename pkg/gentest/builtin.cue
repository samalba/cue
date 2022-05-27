package builtin

// #API defines the exported surface area of a built-in package.
#API: [string]: {
	const: _
} | {
	constraint: #Constraint
} | {
	func: #Func
}

#Constraint: {
	of: _
	args?: [...]
}

#Func: {
	in: [... #Arg]
	out: _
}

#Arg: {
	name: string
	type: _
}

#CommentDecl: {
	// in maps argument name to argument type.
	// Note: this is in a different form to the form used
	// to encode the function type in the generated CUE,
	// as it doesn't have to encode the argument positions,
	// as they're defined by the Go function.
	in: [string]: _
	// out defines the return type.
	out: _
}
