module github.com/ivanizag/izapple2

go 1.12

require (
	fyne.io/fyne v1.4.0-rc1
	github.com/go-gl/glfw/v3.3/glfw v0.0.0-20200625191551-73d3c3675aa3
	github.com/kr/text v0.2.0 // indirect
	github.com/pkg/profile v1.5.0
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749 // indirect
	github.com/shurcooL/vfsgen v0.0.0-20181202132449-6a9ea43bcacd
	github.com/stretchr/testify v1.6.1 // indirect
	github.com/veandco/go-sdl2 v0.4.0
	golang.org/x/net v0.0.0-20201010224723-4f7140c49acb // indirect
	golang.org/x/sys v0.0.0-20201014080544-cc95f250f6bc // indirect
	golang.org/x/tools v0.0.0-20201013201025-64a9e34f3752 // indirect
	gopkg.in/check.v1 v1.0.0-20200902074654-038fdea0a05b // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
)

//replace fyne.io/fyne => github.com/ivanizag/fyne v1.3.4-0.20201010160818-ed5402384cff
//replace fyne.io/fyne => ../fyne/fyne
