package testdata

// 参数1
type HelloParam struct {
	A int `json:"a" form:"a"` // A字段
	B int `json:"b" form:"b"` // B字段
	C HelloParam2
	HelloParam3
}

// 参数2
type HelloParam2 struct {
	A int `json:"a" form:"a"` // A字段2
	B int `json:"b" form:"b"` // B字段2
}

// 参数2
type HelloParam3 struct {
	A2 int `json:"a" form:"a"` // A字段3
	B2 int `json:"b" form:"b"` // B字段3
}
