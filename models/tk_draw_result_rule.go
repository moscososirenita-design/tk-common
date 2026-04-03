package models

// DrawResultZodiacOrder 固定生肖顺序，用于输出集合时保持稳定。
var DrawResultZodiacOrder = []string{"鼠", "牛", "虎", "兔", "龙", "蛇", "马", "羊", "猴", "鸡", "狗", "猪"}

// DrawResultWuxingOrder 固定五行顺序，用于输出集合时保持稳定。
var DrawResultWuxingOrder = []string{"金", "木", "水", "火", "土"}

// DrawResultTailOrder 固定尾数顺序，用于输出集合时保持稳定。
var DrawResultTailOrder = []string{"0尾", "1尾", "2尾", "3尾", "4尾", "5尾", "6尾", "7尾", "8尾", "9尾"}

// DrawResultColorWaveMap 为六合彩波色映射。
var DrawResultColorWaveMap = buildDrawResultStringNumberMap(map[string][]int{
	"红波": {1, 2, 7, 8, 12, 13, 18, 19, 23, 24, 29, 30, 34, 35, 40, 45, 46},
	"蓝波": {3, 4, 9, 10, 14, 15, 20, 25, 26, 31, 36, 37, 41, 42, 47, 48},
	"绿波": {5, 6, 11, 16, 17, 21, 22, 27, 28, 32, 33, 38, 39, 43, 44, 49},
})

// DrawResultZodiacMap 为生肖映射。
var DrawResultZodiacMap = buildDrawResultStringNumberMap(map[string][]int{
	"鼠": {7, 19, 31, 43},
	"牛": {6, 18, 30, 42},
	"虎": {5, 17, 29, 41},
	"兔": {4, 16, 28, 40},
	"龙": {3, 15, 27, 39},
	"蛇": {2, 14, 26, 38},
	"马": {1, 13, 25, 37, 49},
	"羊": {12, 24, 36, 48},
	"猴": {11, 23, 35, 47},
	"鸡": {10, 22, 34, 46},
	"狗": {9, 21, 33, 45},
	"猪": {8, 20, 32, 44},
})

// DrawResultWuxingMap 为五行映射。
var DrawResultWuxingMap = buildDrawResultStringNumberMap(map[string][]int{
	"金": {3, 4, 11, 12, 25, 26, 33, 34, 41, 42},
	"木": {7, 8, 15, 16, 23, 24, 37, 38, 45, 46},
	"水": {13, 14, 21, 22, 29, 30, 43, 44},
	"火": {1, 2, 9, 10, 17, 18, 31, 32, 39, 40, 47, 48},
	"土": {5, 6, 19, 20, 27, 28, 35, 36, 49},
})

// DrawResultBeastMap 为家畜/野兽映射。
var DrawResultBeastMap = map[string]string{
	"猪": "家畜",
	"狗": "家畜",
	"牛": "家畜",
	"马": "家畜",
	"羊": "家畜",
	"鸡": "家畜",
	"鼠": "野兽",
	"虎": "野兽",
	"兔": "野兽",
	"龙": "野兽",
	"蛇": "野兽",
	"猴": "野兽",
}

// buildDrawResultStringNumberMap 将“名称 -> 号码列表”转换成“号码 -> 名称”映射。
func buildDrawResultStringNumberMap(source map[string][]int) map[int]string {
	// 预分配 map，避免多次扩容。
	out := make(map[int]string, 49)
	// 遍历名称和号码列表。
	for label, nums := range source {
		// 逐个回填号码所属的标签。
		for _, num := range nums {
			out[num] = label
		}
	}
	// 返回最终映射。
	return out
}
