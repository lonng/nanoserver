package util

import (
	"errors"
)

// 获取分页编号序列, 页码从 0 开始, 序列中非负整数表示页码, -1 表示省略,
// 如 [0,1,-1,8,9,10,11,12,-1,15,16] 表示 0,1,...,8,9,10,11,12,...,15,16
//  pageNum:          页面数量, 大于 0 的整数
//  currentPageIndex: 当前页码, 从 0 开始编码
func Paginator0(pageNum, currentPageIndex int) ([]int, error) {
	const (
		// 0,1,...,4,5,[6],7,8,...,10,11
		paginatorBeginNum = 2 // 分页开头的页码数量
		paginatorEndNum   = 2 // 分页结尾的页码数量

		currentPageIndexFrontNum  = 2 // 当前页面前面的页码数量
		currentPageIndexBehindNum = 2 // 当前页面后面的页码数量

		currentPageIndexRangeNum = currentPageIndexFrontNum + 1 + currentPageIndexBehindNum // 当前页面游标滑块的页码数量
	)

	if pageNum < 1 {
		return nil, errors.New("pageNum < 1")
	}
	if currentPageIndex < 0 || currentPageIndex >= pageNum {
		return nil, errors.New("currentPageIndex out of range")
	}

	switch {
	case pageNum == 1:
		return []int{0}, nil
	case pageNum <= currentPageIndexRangeNum: // 游标滑块大小内, 肯定是不需要加省略号的
		arr := make([]int, pageNum)
		for i := 0; i < pageNum; i++ {
			arr[i] = i
		}
		return arr, nil
	default: // pageNum > currentPageIndexRangeNum
		maxPageIndex := pageNum - 1

		// 确定当前页面这个游标滑块前后的页码, 如 0,1,...,4,5,[6],7,8,...,10,11 里面的 4 和 8
		rangeBeginPageIndex := currentPageIndex - currentPageIndexFrontNum
		rangeEndPageIndex := currentPageIndex + currentPageIndexBehindNum
		switch {
		case rangeBeginPageIndex < 0:
			rangeBeginPageIndex = 0
			rangeEndPageIndex = currentPageIndexRangeNum - 1 // rangeEndPageIndex == currentPageIndexRangeNum - 1 < pageNum -1 == maxPageIndex
		case rangeEndPageIndex > maxPageIndex:
			rangeEndPageIndex = maxPageIndex
			rangeBeginPageIndex = maxPageIndex - (currentPageIndexRangeNum - 1) // rangeBeginPageIndex == maxPageIndex - (currentPageIndexRangeNum - 1) == pageNum -1 - (currentPageIndexRangeNum - 1) == pageNum - currentPageIndexRangeNum > 0
		}

		if rangeBeginPageIndex <= paginatorBeginNum { // 跟前面相连
			if rangeEndPageIndex >= maxPageIndex-paginatorEndNum { // 跟后面相连
				arr := make([]int, pageNum)
				for i := 0; i < pageNum; i++ {
					arr[i] = i
				}
				return arr, nil
			} else { //跟后面不连
				arr := make([]int, 0, rangeEndPageIndex+1+1+paginatorEndNum)
				for i := 0; i <= rangeEndPageIndex; i++ {
					arr = append(arr, i)
				}
				arr = append(arr, -1)
				for i := pageNum - paginatorEndNum; i < pageNum; i++ {
					arr = append(arr, i)
				}
				return arr, nil
			}
		} else { // 跟前面不连
			if rangeEndPageIndex >= maxPageIndex-paginatorEndNum { // 跟后面相连
				arr := make([]int, 0, paginatorBeginNum+1+(pageNum-rangeBeginPageIndex))
				for i := 0; i < paginatorBeginNum; i++ {
					arr = append(arr, i)
				}
				arr = append(arr, -1)
				for i := rangeBeginPageIndex; i < pageNum; i++ {
					arr = append(arr, i)
				}
				return arr, nil
			} else { //跟后面不连
				arr := make([]int, 0, paginatorBeginNum+1+currentPageIndexRangeNum+1+paginatorEndNum)
				for i := 0; i < paginatorBeginNum; i++ {
					arr = append(arr, i)
				}
				arr = append(arr, -1)
				for i := rangeBeginPageIndex; i <= rangeEndPageIndex; i++ {
					arr = append(arr, i)
				}
				arr = append(arr, -1)
				for i := pageNum - paginatorEndNum; i < pageNum; i++ {
					arr = append(arr, i)
				}
				return arr, nil
			}
		}
	}
}

// 获取分页编号序列, 页码从 1 开始, 序列中正整数表示页码, -1 表示省略,
// 如 [1,2,-1,8,9,10,11,12,-1,15,16] 表示 1,2,...,8,9,10,11,12,...,15,16
//  pageNum:          页面数量, 大于 0 的整数
//  currentPageIndex: 当前页码, 从 1 开始编码
func Paginator1(pageNum, currentPageIndex int) (arr []int, err error) {
	currentPageIndex--
	arr, err = Paginator0(pageNum, currentPageIndex)
	if err != nil {
		return
	}
	for i := 0; i < len(arr); i++ {
		if arr[i] != -1 {
			arr[i]++
		}
	}
	return
}

// 获取分页编号序列, 页码从 0 开始, 序列中非负整数表示页码, -1 表示省略,
// 如 [0,1,-1,8,9,10,11,12,-1,15,16] 表示 0,1,...,8,9,10,11,12,...,15,16
//  totalItemNum:     总的记录数量, 不是页面数量, 非负整数
//  pageSize:         每页显示的数量, 大于 0 的整数
//  currentPageIndex: 当前页码, 从 0 开始编码
func Paginator0Ex(totalItemNum, pageSize, currentPageIndex int) (arr []int, pageNum int, err error) {
	if totalItemNum < 0 {
		err = errors.New("totalItemNum < 0")
		return
	}
	if pageSize <= 0 {
		err = errors.New("pageSize <= 0")
		return
	}

	if totalItemNum == 0 {
		pageNum = 1
	} else {
		pageNum = (totalItemNum + pageSize - 1) / pageSize
	}

	arr, err = Paginator0(pageNum, currentPageIndex)
	return
}

// 获取分页编号序列, 页码从 1 开始, 序列中正整数表示页码, -1 表示省略,
// 如 [1,2,-1,8,9,10,11,12,-1,15,16] 表示 1,2,...,8,9,10,11,12,...,15,16
//  totalItemNum:     总的记录数量, 不是页面数量, 非负整数
//  pageSize:         每页显示的数量, 大于 0 的整数
//  currentPageIndex: 当前页码, 从 1 开始编码
func Paginator1Ex(totalItemNum, pageSize, currentPageIndex int) (arr []int, pageNum int, err error) {
	if totalItemNum < 0 {
		err = errors.New("totalItemNum < 0")
		return
	}
	if pageSize <= 0 {
		err = errors.New("pageSize <= 0")
		return
	}

	if totalItemNum == 0 {
		pageNum = 1
	} else {
		pageNum = (totalItemNum + pageSize - 1) / pageSize
	}

	arr, err = Paginator1(pageNum, currentPageIndex)
	return
}
