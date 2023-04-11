package sortedset

import (
	"errors"
	"strconv"
)

/*
 * ScoreBorder 是一个代表Redis命令 `ZRANGEBYSCORE`的最大值（max）和最小值（min）的结构体
 * can accept:
 *   int or float value, such as 2.718, 2, -2.718, -2 ...
 *   exclusive int or float value, such as (2.718, (2, (-2.718, (-2 ...
 *   infinity: +inf, -inf， inf(same as +inf)
 */

const (
	negativeInf int8 = -1 // 负无穷
	positiveInf int8 = 1  // 正无穷
)

// ScoreBorder 代表一个浮点值的边界范围: <, <=, >, >=, +inf, -inf，例如>=3、<6、(0,正无穷)等
type ScoreBorder struct {
	Inf     int8    //表示范围的极限
	Value   float64 // 数值的具体取值
	Exclude bool    // 是否需要排除在范围之外，true表示不在范围内，是开区间，false表示不排除，是闭区间
}

// if max.greater(score) then the score is within the upper border
// do not use min.greater()
func (border *ScoreBorder) greater(value float64) bool {
	if border.Inf == negativeInf {
		// 如果表示负无穷，则肯定比value小，所以返回false
		return false
	} else if border.Inf == positiveInf {
		// 如果表示正无穷，则肯定比value大，所以返回true
		return true
	}
	// 判断是否大于等于
	if border.Exclude {
		return border.Value > value
	}
	return border.Value >= value
}

// 与
func (border *ScoreBorder) less(value float64) bool {
	if border.Inf == negativeInf {
		return true
	} else if border.Inf == positiveInf {
		return false
	}
	if border.Exclude {
		return border.Value < value
	}
	return border.Value <= value
}

var positiveInfBorder = &ScoreBorder{
	Inf: positiveInf,
}

var negativeInfBorder = &ScoreBorder{
	Inf: negativeInf,
}

// ParseScoreBorder 将Redis的参数解析为 ScoreBorder
func ParseScoreBorder(s string) (*ScoreBorder, error) {
	if s == "inf" || s == "+inf" {
		return positiveInfBorder, nil
	}
	if s == "-inf" {
		return negativeInfBorder, nil
	}
	// (0
	if s[0] == '(' {
		value, err := strconv.ParseFloat(s[1:], 64)
		if err != nil {
			return nil, errors.New("ERR min or max is not a float")
		}
		return &ScoreBorder{
			Inf:     0,
			Value:   value,
			Exclude: true,
		}, nil
	}
	// 3)
	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, errors.New("ERR min or max is not a float")
	}
	return &ScoreBorder{
		Inf:     0,
		Value:   value,
		Exclude: false,
	}, nil
}
