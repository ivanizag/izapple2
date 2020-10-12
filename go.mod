module github.com/ivanizag/izapple2

go 1.12

require (
	fyne.io/fyne v1.3.3
	github.com/go-gl/glfw/v3.3/glfw v0.0.0-20200625191551-73d3c3675aa3
	github.com/pkg/profile v1.4.0
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749 // indirect
	github.com/shurcooL/vfsgen v0.0.0-20181202132449-6a9ea43bcacd
	github.com/veandco/go-sdl2 v0.4.0
)

//replace fyne.io/fyne => github.com/ivanizag/fyne v1.3.4-0.20201010160818-ed5402384cff
// replace fyne.io/fyne => ../../fyne/fyne
