package common

// 请求参数类型定义 -- 公共类型
// json字段统一使用大驼峰命名


// 分页参数
type Page struct {
	PageNum int `json:"PageNum" min:"1" max:"100" default:"1"`
	PageSize int `json:"PageSize" min:"1" max:"100" default:"10"`
}
