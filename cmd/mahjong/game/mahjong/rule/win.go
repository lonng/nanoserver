package rule

import (
	"fmt"
	"sort"

	"github.com/lonnng/nanoserver/cmd/mahjong/game/mahjong"
)

//http://blog.csdn.net/loushuai/article/details/51785761
//http://www.xqbase.com/other/mahjongg_english.htm

var result mahjong.Result

var debug bool = false

func pushResult(args ...int) {
	result = append(result, args...)
}

//eraseValue 移除一个指定值
func eraseValue(in []int, val int) []int {
	for i, v := range in {
		if v == val {
			return append(in[:i], in[i+1:]...)
		}
	}
	return in
}

//Shrink 剪枝
func Shrink(in []int) ([]int, bool) {
	if len(in) < 3 {
		return in, false
	}
	if debug {
		fmt.Println("enter shrink with data:", in)
	}

	//len(A)== 3,4
	if in[0] == in[1] && in[1] == in[2] {
		pushResult(in[0], in[1], in[2])

		begin := in[0]
		if debug {
			fmt.Printf("此分组内的最小元素:%d至少有3个，移除3个%d后的结果是:%+v %d \n", begin, begin, in[3:], len(in[3:]))
		}
		return in[3:], true

		//len(A)== 1, 2
	} else {
		if in[len(in)-1]-in[0] < 2 {
			if debug {
				fmt.Printf("由于每次需要移除3个元素,而处于当前情况下，此分组的最小元素:%d 最多有2个,要满足移除条件 A, A+1, A+2,则此序列的最后一个元素不能小于%d 但实际值为: %d\n",
					in[0],
					in[0]+2,
					in[len(in)-1])
			}
			return nil, false
		}

	}

	var (
		i         int
		t         = in[0]
		beginWith = t
	)

	for i < 3 {
		if debug {
			fmt.Printf("移除以%d开始的顺子序列中的元素: %d\n", beginWith, t)
		}
		in = eraseValue(in, t)
		i++
		t++
	}

	pushResult(t-3, t-2, t-1)
	return in, true
}

func Group(input []int) [][]int {
	if debug {
		fmt.Printf("进入 group with data:%+v\n", input)
	}

	if len(input) == 0 {
		if debug {
			fmt.Println("已经无更多元素需要分组,退出分组")
		}
		return nil
	}

	member := []int{}
	group := make([][]int, 0)

	beginWith := input[0]
	member = append(member, input[0])
	if debug {
		fmt.Printf("以此组的第一个元素: %d 为首进行分组\n", beginWith)
	}

	//for k,v := range input[1:] {
	for k := 1; k < len(input); k++ {
		v := input[k]
		if v == input[k-1] || v == input[k-1]+1 {
			member = append(member, v)
			if debug {
				fmt.Printf("%d与%d是连续的 分为一组\n", v, input[k-1])
			}
			continue
		}
		group = append(group, member)

		if debug {
			fmt.Printf("以%d为首的组,已经构建完成 %+v\n", beginWith, group)
		}

		member = []int{}
		beginWith = v

		if debug {
			fmt.Printf("\n开始以%d为首的新组\n", beginWith)
		}
		member = append(member, v)
	}
	if debug {
		fmt.Println("将最后一个分组入队")
	}
	group = append(group, member)
	return group
}

func IsLegal(onHand []int) bool {

	if len(onHand) == 0 {
		if debug {
			fmt.Println("已经无更多元素组需要检测,退出")
		}
		return true
	}
	if debug {
		fmt.Println("\n\n进入isLegal with data:", onHand)
	}
	if len(onHand)%3 != 0 {
		if debug {
			fmt.Println("除去对子后，余下的元数的总个数必须是3的倍数,而实际是: ", len(onHand))
		}
		return false
	}

	if len(onHand) == 3 {
		if debug {
			fmt.Println("此组元素队列只有3个元素")
		}
		if onHand[0] == onHand[1] && onHand[1] == onHand[2] {
			if debug {
				fmt.Println("此组元素是刻子: ", onHand[0], onHand[1], onHand[2])
			}

			pushResult(onHand[0], onHand[1], onHand[2])
			return true
		}
		if onHand[1] == onHand[0]+1 && onHand[2] == onHand[1]+1 {
			pushResult(onHand[0], onHand[1], onHand[2])

			if debug {
				fmt.Println("此组元素是顺子: ", onHand[0], onHand[1], onHand[2])
			}
			return true
		}

		if debug {
			fmt.Println("此组元素是不满足(A,A,A)| (A, A+1, A+2)这个性质: ", onHand[0], onHand[1], onHand[2])
		}
		return false
	}

	if debug {
		fmt.Println("此组元素多于3个，根据连续性( AAA | A, A+1, A+2)对其进行分组")
	}
	grp := Group(onHand)
	if debug {
		fmt.Println("分组列表:", grp)
		fmt.Println("分组完成,准备对每一组进行剪枝")
	}

	var (
		remained []int
		ok       bool
	)
	for i, row := range grp {
		if debug {
			fmt.Printf("分组%d剪枝前: %+v\n", i, row)
		}
		if remained, ok = Shrink(row); !ok {
			if debug {
				fmt.Println("剪枝失败,未通过测试,返回上一层")
			}
			return false
		}
		if debug {
			fmt.Println("剪枝成功,对剪枝后的序列,继续合法性判断", len(remained), remained)
		}
		if !IsLegal(remained) {
			if debug {
				fmt.Println("当前分支的子分支,未通过合法性判断,,返回上一层")
			}
			return false
		}
	}

	if debug {
		fmt.Println("当前分组(及子分组)均已经检测完毕,检测通过")
	}
	return true
}

//是否和牌
func IsWin(mj mahjong.Indexes) bool {
	return IsWinWithIndexes(mj)
}

func IsWinWithIndexes(indexes mahjong.Indexes) bool {
	sort.Ints(indexes)

	result = nil
	var prevPair int

	i := 0
	for i < len(indexes)-1 {
		result = result[:0]

		if indexes[i] != indexes[i+1] || indexes[i] == prevPair {
			i++
			continue
		}

		origin := make([]int, len(indexes))
		copy(origin, indexes)
		if debug {
			fmt.Println("原始序列的copy: ", i, origin)
		}

		if indexes[i] == indexes[i+1] {
			pushResult(indexes[i : i+2]...)
			prevPair = indexes[i]
			if debug {
				fmt.Println("找到对子: ", indexes[i])
				fmt.Println("移除对子前: ", origin)
			}
			origin = append(origin[:i], origin[i+2:]...)
			if debug {
				fmt.Println("移除对子后: ", origin)
			}

			if IsLegal(origin) {
				return true
			}
		}
		i++

	}

	return false
}

func EnableDebug(enable bool) {
	debug = enable
}
