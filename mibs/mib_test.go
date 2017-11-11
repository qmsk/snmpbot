package mibs

var (
	TestMIB = RegisterMIB("TEST-MIB", 1, 0, 1)

	TestObject = TestMIB.RegisterObject(TestMIB.MakeID("test", 1, 1), Object{
		Syntax: DisplayStringSyntax{},
	})
)
