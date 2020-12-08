package file

//定义一个通用的结构体
type Bucket struct {
	Slice []FileInfoX                 //承载以任意结构体为元素构成的Slice
	By    func(a, b interface{}) bool //排序规则函数,当需要对新的结构体slice进行排序时，只需定义这个函数即可
}

/*
定义三个必须方法的准则：接收者不能为指针
*/
func (this Bucket) Len() int { return len(this.Slice) }

func (this Bucket) Swap(i, j int) { this.Slice[i], this.Slice[j] = this.Slice[j], this.Slice[i] }

func (this Bucket) Less(i, j int) bool { return this.By(this.Slice[i], this.Slice[j]) }
