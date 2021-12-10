package configs

// Evaluation form status
const (
	EvaluationMemberIsCreatingStatus     = 1
	EvaluationCreatedStatus              = 2
	EvaluationMemberIsEditingStatus      = 3
	EvaluationMemberEditedStatus         = 4
	EvaluationVNManagerIsReviewingStatus = 5
	EvaluationVNManagerReviewedStatus    = 6
)

// EvaluationStatus map
var EvaluationStatus = map[int]string{
	1: "Member is creating",
	2: "Created",
	3: "Member is editing",
	4: "Member edited",
	5: "VN manager is reviewing",
	6: "VN manager reviewed",
}

var EvaluationCategoriesEN = map[string]string{
	"CT1":  "Goal management",
	"CT2":  "Year",
	"CT3":  "Quarter",
	"CT4":  "■ Result",
	"CT5":  "○ Final result",
	"CT6":  "Full name",
	"CT7":  "Rank",
	"CT8":  "Goal table",
	"CT9":  "Position",
	"CT10": "Total score",
	"CT11": "Result",
	"CT12": "Rank",
	"CT13": "Goal table",
	"CT14": "Position",
	"CT15": "・・・Self assessment",
	"CT16": "Department",
	"CT17": "・・・Supervisor assessment",
	"CT18": "Total weight must be equal 100",
	"CT19": "Total weight",
	"CT20": "■ General objective",
	"CT21": "○ Please confirm with your supervisor before entering the goal value",
	"CT22": "Weight",
	"CT23": "Item",
	"CT24": "Goal value\n(unit is not required)",
	"CT25": "Real value\n(unit is not required)",
	"CT26": "Rate achieved",
	"CT27": "Score",
	"CT28": "■ Specific objective",
	"CT29": "○ Please set goals both quantitatively and qualitatively",
	"CT30": "Weight",
	"CT31": "Item (quantitatively)",
	"CT32": "Goal value\n(unit is not required)",
	"CT33": "Real value\n(unit is not required)",
	"CT34": "Rate achieved",
	"CT35": "Score",
	"CT36": "Weight",
	"CT37": "Item (qualitatively)",
	"CT38": "Detail\n(Please write according to 5W1H pattern)",
	"CT39": "Rate achieved\n(Self assessment)",
	"CT40": "Rate achieved\n(Supervisor assessment)",
	"CT41": "Score",
	"CT42": "■ Comments",
	"CT43": "○ Self comment",
	"CT44": "○ Supervisor comment",
}

var EvaluationCategoriesJP = map[string]string{
	"CT1":  "目標管理シート",
	"CT2":  "年 第",
	"CT3":  "四半期",
	"CT4":  "■評価結果",
	"CT5":  "○評価結果は、評価確認会議で最終決定します。",
	"CT6":  "氏名",
	"CT7":  "等級・ランク",
	"CT8":  "評価ﾃｰﾌﾞﾙ",
	"CT9":  "役割給",
	"CT10": "評価点合計",
	"CT11": "評価結果",
	"CT12": "等級・ランク",
	"CT13": "評価テーブル",
	"CT14": "役割給",
	"CT15": "・・・本人入力欄",
	"CT16": "部署",
	"CT17": "・・・上長入力欄",
	"CT18": "目標の「ウェイト合計が100」になるように設定して下さい。",
	"CT19": "ウェイト合計",
	"CT20": "■共通目標",
	"CT21": "○共通目標の目標・実績数値は、上長に確認して入力して下さい。",
	"CT22": "ウェイト",
	"CT23": "評価指標",
	"CT24": "目標数値\n（単位は入力不要です）",
	"CT25": "実績数値\n（単位は入力不要です）",
	"CT26": "達成率",
	"CT27": "評価点",
	"CT28": "■個別目標",
	"CT29": "○数値目標は上部の欄に、取組課題目標は下部の欄に入力して下さい。",
	"CT30": "ウェイト",
	"CT31": "評価指標（数値目標）",
	"CT32": "目標数値\n（単位は入力不要です）",
	"CT33": "実績数値\n（単位は入力不要です）",
	"CT34": "達成率",
	"CT35": "評価点",
	"CT36": "ウェイト",
	"CT37": "評価指標（取組課題目標）",
	"CT38": "行動項目\n（5W1Hで記載して下さい）",
	"CT39": "達成率\n自己評価",
	"CT40": "達成率\n上長",
	"CT41": "評価点",
	"CT42": "■評価時コメント記入欄",
	"CT43": "○自己",
	"CT44": "○上長",
}

var EvaluationCategoriesVN = map[string]string{
	"CT1":  "Quản lý mục tiêu",
	"CT2":  "Năm",
	"CT3":  "Quý",
	"CT4":  "■ Kết quả đánh giá",
	"CT5":  "○ Kết quả đánh giá cuối cùng",
	"CT6":  "Họ tên",
	"CT7":  "Cấp bậc",
	"CT8":  "Bảng đánh giá",
	"CT9":  "Chức vụ",
	"CT10": "Tổng điểm",
	"CT11": "Kết quả",
	"CT12": "Cấp bậc",
	"CT13": "Bảng đánh giá",
	"CT14": "Chức vụ",
	"CT15": "・・・Phần đánh giá của mình",
	"CT16": "Chi nhánh",
	"CT17": "・・・Phần đánh giá của cấp trên",
	"CT18": "Hãy thiết lập tổng trọng số phải bằng 100",
	"CT19": "Tổng trọng số",
	"CT20": "■ Mục tiêu chung",
	"CT21": "○ Hãy xác nhận với cấp trên trước khi điền giá trị mục tiêu",
	"CT22": "Trọng số",
	"CT23": "Chỉ mục đánh giá",
	"CT24": "Mục tiêu bằng con số\n(không cần nhập đơn vị)",
	"CT25": "Thực tế đạt được bằng con số\n(không cần nhập đơn vị)",
	"CT26": "Tỷ lệ đạt được",
	"CT27": "Điểm đánh giá",
	"CT28": "■ Mục tiêu cá nhân",
	"CT29": "○ Hãy thiết lập mục tiêu có cả phần định lượng và định tính",
	"CT30": "Trọng số",
	"CT31": "Chỉ mục đánh giá(định lượng)",
	"CT32": "Mục tiêu bằng con số\n(không cần nhập đơn vị)",
	"CT33": "Thực tế đạt được bằng con số\n(không cần nhập đơn vị)",
	"CT34": "Tỷ lệ đạt được",
	"CT35": "Điểm đánh giá",
	"CT36": "Trọng số",
	"CT37": "Chỉ mục đánh giá(định tính)",
	"CT38": "Chi tiết\n(hãy viết mục tiêu theo tiêu chí 5W1H)",
	"CT39": "Tỷ lệ hoàn thành\n(Tự đánh giá)",
	"CT40": "Tỷ lệ hoàn thành\n(Cấp trên)",
	"CT41": "Điểm đánh giá",
	"CT42": "■ Nhận xét bản đánh giá",
	"CT43": "○ Tự nhận xét",
	"CT44": "○ Cấp trên nhận xét",
}

var ExcelStyles = map[string]string{
	"fillBackgroundColor": `{
		"border":[
			{"type":"bottom","color":"#FFFFFF","style":1},
			{"type":"left","color":"#FFFFFF","style":1},
			{"type":"right","color":"#FFFFFF","style":1},
			{"type":"top","color":"#FFFFFF","style":1}
		]
	}`,
	"bold": `{"font":{"bold":true, "size":13}}`,
	"topDotBorder": `{
		"border":[
			{"type":"bottom","color":"#FFFFFF","style":1},
			{"type":"left","color":"#FFFFFF","style":1},
			{"type":"right","color":"#FFFFFF","style":1},
			{"type":"top","color":"#000000","style":7}
		]
	}`,
	"allThinBorderCenter": `{
		"font":{"color":"#3333CC"},
		"fill":{"type":"pattern","color":["#FFFF99"],"pattern":1},
		"alignment":{"horizontal":"center", "vertical":"center"},
		"border":[
			{"type":"bottom","color":"#000000","style":1},
			{"type":"top","color":"#000000","style":1},
			{"type":"right","color":"#000000","style":1},
			{"type":"left","color":"#000000","style":1}
		]
	}`,
	"leftBorderBold": `{
		"border":[
			{"type":"left","color":"#000000","style":2},
			{"type":"bottom","color":"#FFFFFF","style":1},
			{"type":"right","color":"#FFFFFF","style":1},
			{"type":"top","color":"#FFFFFF","style":1}
		]
	}`,
	"topBorderBold": `{
		"border":[
			{"type":"top","color":"#000000","style":2},
			{"type":"bottom","color":"#FFFFFF","style":1},
			{"type":"left","color":"#FFFFFF","style":1},
			{"type":"right","color":"#FFFFFF","style":1}
		]
	}`,
	"bottomBorderBold": `{
		"border":[
			{"type":"bottom","color":"#000000","style":2},
			{"type":"left","color":"#FFFFFF","style":1},
			{"type":"right","color":"#FFFFFF","style":1},
			{"type":"top","color":"#FFFFFF","style":1}
		]
	}`,
	"rightBorderBold": `{
		"border":[
			{"type":"right","color":"#000000","style":2},
			{"type":"bottom","color":"#FFFFFF","style":1},
			{"type":"left","color":"#FFFFFF","style":1},
			{"type":"top","color":"#FFFFFF","style":1}
		]
	}`,
	"rightAndLeftBorderBold": `{
		"border":[
			{"type":"right","color":"#000000","style":2},
			{"type":"left","color":"#000000","style":2},
			{"type":"bottom","color":"#FFFFFF","style":1},
			{"type":"top","color":"#FFFFFF","style":1}
		]
	}`,
	"topAndBottomBorderBold": `{
		"border":[
			{"type":"top","color":"#000000","style":2},
			{"type":"bottom","color":"#000000","style":2},
			{"type":"left","color":"#FFFFFF","style":1},
			{"type":"right","color":"#FFFFFF","style":1}
		]
	}`,
	"leftAndBottomBorderThinFontBold": `{
		"font":{"bold":true},
		"alignment":{"horizontal":"center", "vertical":"center"},
		"border":[
			{"type":"left","color":"#000000","style":1},
			{"type":"bottom","color":"#000000","style":1},
			{"type":"right","color":"#FFFFFF","style":1},
			{"type":"top","color":"#FFFFFF","style":1}
		]
	}`,
	"topRightBoldLeftBottomThinBorderFontBold": `{
		"font":{"bold":true},
		"alignment":{"horizontal":"center", "vertical":"center"},
		"border":[
			{"type":"left","color":"#000000","style":1},
			{"type":"bottom","color":"#000000","style":1},
			{"type":"top","color":"#000000","style":2},
			{"type":"right","color":"#000000","style":2}
		]
	}`,
	"rightAndBottomBorderThinFontBold": `{
		"font":{"bold":true},
		"alignment":{"horizontal":"center", "vertical":"center"},
		"border":[
			{"type":"right","color":"#000000","style":1},
			{"type":"bottom","color":"#000000","style":1},
			{"type":"left","color":"#FFFFFF","style":1},
			{"type":"top","color":"#FFFFFF","style":1}
		]
	}`,
	"topLeftBoldRightBottomBorderThinFontBold": `{
		"font":{"bold":true},
		"alignment":{"horizontal":"center", "vertical":"center"},
		"border":[
			{"type":"right","color":"#000000","style":1},
			{"type":"bottom","color":"#000000","style":1},
			{"type":"top","color":"#000000","style":2},
			{"type":"left","color":"#000000","style":2}
		]
	}`,
	"topBoldRightBottomBorderThinFontBold": `{
		"font":{"bold":true},
		"alignment":{"horizontal":"center", "vertical":"center"},
		"border":[
			{"type":"right","color":"#000000","style":1},
			{"type":"bottom","color":"#000000","style":1},
			{"type":"top","color":"#000000","style":2},
			{"type":"left","color":"#FFFFFF","style":1}
		]
	}`,
	"topAndRightBorderThin": `{
		"alignment":{"horizontal":"center", "vertical":"center"},
		"border":[
			{"type":"top","color":"#000000","style":1},
			{"type":"right","color":"#000000","style":1},
			{"type":"left","color":"#FFFFFF","style":1}
		]
	}`,
	"topAndRightBorderThinFontBold": `{
		"font":{"bold":true},
		"alignment":{"horizontal":"center", "vertical":"center"},
		"border":[
			{"type":"top","color":"#000000","style":1},
			{"type":"right","color":"#000000","style":1},
			{"type":"bottom","color":"#FFFFFF","style":1},
			{"type":"left","color":"#FFFFFF","style":1}
		]
	}`,
	"topAndLeftBorderThin": `{
		"alignment":{"horizontal":"center", "vertical":"center"},
		"border":[
			{"type":"top","color":"#000000","style":1},
			{"type":"left","color":"#000000","style":1},
			{"type":"right","color":"#FFFFFF","style":1}
		]
	}`,
	"leftBorderThin": `{
		"alignment":{"horizontal":"center", "vertical":"center"},
		"border":[
			{"type":"left","color":"#000000","style":1},
			{"type":"bottom","color":"#FFFFFF","style":1},
			{"type":"right","color":"#FFFFFF","style":1},
			{"type":"top","color":"#FFFFFF","style":1}
		]
	}`,
	"topBorderThin": `{
		"alignment":{"horizontal":"center", "vertical":"center"},
		"border":[
			{"type":"top","color":"#000000","style":1},
			{"type":"bottom","color":"#FFFFFF","style":1},
			{"type":"left","color":"#FFFFFF","style":1},
			{"type":"right","color":"#FFFFFF","style":1}
		]
	}`,
	"topDotLeftContinuousBorder": `{
		"alignment":{"horizontal":"center", "vertical":"center"},
		"border":[
			{"type":"top","color":"#000000","style":7},
			{"type":"left","color":"#000000","style":1},
			{"type":"right","color":"#FFFFFF","style":1}
		]
	}`,
	"allThinBorderUser": `{
		"font":{"color":"#3333CC"},
		"fill":{"type":"pattern","color":["#FFFF99"],"pattern":1},
		"alignment":{"horizontal":"left", "vertical":"top"},
		"border":[
			{"type":"bottom","color":"#000000","style":1},
			{"type":"left","color":"#000000","style":1},
			{"type":"right","color":"#000000","style":1},
			{"type":"top","color":"#000000","style":1}
		]
	}`,
	"allBoldBorderUser": `{
		"font":{"color":"#3333CC"},
		"fill":{"type":"pattern","color":["#FFFF99"],"pattern":1},
		"alignment":{"horizontal":"left", "vertical":"top"},
		"border":[
			{"type":"bottom","color":"#000000","style":2},
			{"type":"left","color":"#000000","style":2},
			{"type":"right","color":"#000000","style":2},
			{"type":"top","color":"#000000","style":2}
		]
	}`,
	"rightAndBottomBorderBoldUser": `{
		"font":{"color":"#3333CC"},
		"fill":{"type":"pattern","color":["#FFFF99"],"pattern":1},
		"alignment":{"horizontal":"center", "vertical":"center"},
		"border":[
			{"type":"right","color":"#000000","style":2},
			{"type":"bottom","color":"#000000","style":2},
			{"type":"left","color":"#FFFFFF","style":1},
			{"type":"top","color":"#FFFFFF","style":1}
		]
	}`,
	"leftAndBottomBorderBoldUser": `{
		"font":{"color":"#3333CC"},
		"fill":{"type":"pattern","color":["#FFFF99"],"pattern":1},
		"alignment":{"horizontal":"center", "vertical":"center"},
		"border":[
			{"type":"left","color":"#000000","style":2},
			{"type":"bottom","color":"#000000","style":2},
			{"type":"right","color":"#FFFFFF","style":1},
			{"type":"top","color":"#FFFFFF","style":1}
		]
	}`,
	"rightAndBottomBorderThinUser": `{
		"font":{"color":"#3333CC"},
		"fill":{"type":"pattern","color":["#FFFF99"],"pattern":1},
		"alignment":{"horizontal":"center", "vertical":"center"},
		"border":[
			{"type":"right","color":"#000000","style":1},
			{"type":"bottom","color":"#000000","style":1},
			{"type":"left","color":"#FFFFFF","style":1},
			{"type":"top","color":"#FFFFFF","style":1}
		]
	}`,
	"bottomDotTopAndRightBorderThinUser": `{
		"font":{"color":"#3333CC"},
		"fill":{"type":"pattern","color":["#FFFF99"],"pattern":1},
		"alignment":{"horizontal":"center", "vertical":"center"},
		"border":[
			{"type":"top","color":"#000000","style":1},
			{"type":"right","color":"#000000","style":1},
			{"type":"left","color":"#FFFFFF","style":1},
			{"type":"bottom","color":"#000000","style":7}
		]
	}`,
	"bottomDotTopAndRightHorizontalLeftBorderThinUser": `{
		"font":{"color":"#3333CC"},
		"fill":{"type":"pattern","color":["#FFFF99"],"pattern":1},
		"alignment":{"horizontal":"left", "vertical":"center"},
		"border":[
			{"type":"top","color":"#000000","style":1},
			{"type":"right","color":"#000000","style":1},
			{"type":"left","color":"#FFFFFF","style":1},
			{"type":"bottom","color":"#000000","style":7}
		]
	}`,
	"leftBorderThinUser": `{
		"font":{"color":"#3333CC"},
		"fill":{"type":"pattern","color":["#FFFF99"],"pattern":1},
		"alignment":{"horizontal":"center", "vertical":"center"},
		"border":[
			{"type":"left","color":"#000000","style":1},
			{"type":"bottom","color":"#FFFFFF","style":1},
			{"type":"right","color":"#FFFFFF","style":1},
			{"type":"top","color":"#FFFFFF","style":1}
		]
	}`,
	"topBorderThinUser": `{
		"font":{"color":"#3333CC"},
		"fill":{"type":"pattern","color":["#FFFF99"],"pattern":1},
		"alignment":{"horizontal":"center", "vertical":"center"},
		"border":[
			{"type":"top","color":"#000000","style":1},
			{"type":"bottom","color":"#FFFFFF","style":1},
			{"type":"left","color":"#FFFFFF","style":1},
			{"type":"right","color":"#FFFFFF","style":1}
		]
	}`,
	"topDotRightContinuousBorderUser": `{
		"font":{"color":"#3333CC"},
		"fill":{"type":"pattern","color":["#FFFF99"],"pattern":1},
		"alignment":{"horizontal":"center", "vertical":"center"},
		"border":[
			{"type":"top","color":"#000000","style":7},
			{"type":"right","color":"#000000","style":1},
			{"type":"left","color":"#FFFFFF","style":1},
			{"type":"bottom","color":"#000000","style":7}
		]
	}`,
	"topLeftRightContinuousBorderUser": `{
		"font":{"color":"#3333CC"},
		"fill":{"type":"pattern","color":["#FFFF99"],"pattern":1},
		"alignment":{"horizontal":"center", "vertical":"center"},
		"border":[
			{"type":"top","color":"#000000","style":1},
			{"type":"right","color":"#000000","style":1},
			{"type":"left","color":"#000000","style":1},
		]
	}`,
	"topBottomDotRightContinuousHorizontalLeftBorderUser": `{
		"font":{"color":"#3333CC"},
		"fill":{"type":"pattern","color":["#FFFF99"],"pattern":1},
		"alignment":{"horizontal":"left", "vertical":"center"},
		"border":[
			{"type":"top","color":"#000000","style":7},
			{"type":"right","color":"#000000","style":1},
			{"type":"left","color":"#FFFFFF","style":1},
			{"type":"bottom","color":"#FFFFFF","style":7}
		]
	}`,
	"bottomBoldRightLeftThinBorderUser": `{
		"font":{"color":"#3333CC"},
		"fill":{"type":"pattern","color":["#FFFF99"],"pattern":1},
		"alignment":{"horizontal":"center", "vertical":"center"},
		"border":[
			{"type":"bottom","color":"#000000","style":2},
			{"type":"left","color":"#000000","style":1},
			{"type":"right","color":"#000000","style":1},
			{"type":"top","color":"#FFFFFF","style":1}
		]
	}`,
	"allBoldBorderSupervisor": `{
		"font":{"color":"#3333CC"},
		"fill":{"type":"pattern","color":["#99CCFF"],"pattern":1},
		"alignment":{"horizontal":"left", "vertical":"top"},
		"border":[
			{"type":"right","color":"#000000","style":2},
			{"type":"bottom","color":"#000000","style":2},
			{"type":"top","color":"#000000","style":2},
			{"type":"left","color":"#000000","style":2}
		]
	}`,
	"allThinBorderSupervisor": `{
		"font":{"color":"#3333CC"},
		"fill":{"type":"pattern","color":["#99CCFF"],"pattern":1},
		"alignment":{"horizontal":"left", "vertical":"top"},
		"border":[
			{"type":"right","color":"#000000","style":1},
			{"type":"bottom","color":"#000000","style":1},
			{"type":"top","color":"#000000","style":1},
			{"type":"left","color":"#000000","style":1}
		]
	}`,
	"rightAndBottomBorderBoldSupervisor": `{
		"font":{"color":"#3333CC"},
		"fill":{"type":"pattern","color":["#99CCFF"],"pattern":1},
		"alignment":{"horizontal":"center", "vertical":"center"},
		"border":[
			{"type":"right","color":"#000000","style":2},
			{"type":"bottom","color":"#000000","style":2},
			{"type":"left","color":"#FFFFFF","style":1},
			{"type":"top","color":"#FFFFFF","style":1}
		]
	}`,
	"leftAndBottomBorderBoldSupervisor": `{
		"font":{"color":"#3333CC"},
		"fill":{"type":"pattern","color":["#99CCFF"],"pattern":1},
		"alignment":{"horizontal":"center", "vertical":"center"},
		"border":[
			{"type":"left","color":"#000000","style":2},
			{"type":"bottom","color":"#000000","style":2},
			{"type":"right","color":"#FFFFFF","style":1},
			{"type":"top","color":"#FFFFFF","style":1}
		]
	}`,
	"topAndRightBorderThinSupervisor": `{
		"font":{"color":"#3333CC"},
		"fill":{"type":"pattern","color":["#99CCFF"],"pattern":1},
		"alignment":{"horizontal":"center", "vertical":"center"},
		"border":[
			{"type":"top","color":"#000000","style":1},
			{"type":"right","color":"#000000","style":1},
			{"type":"left","color":"#FFFFFF","style":1}
		]
	}`,
	"leftBorderThinSupervisor": `{
		"font":{"color":"#3333CC"},
		"fill":{"type":"pattern","color":["#99CCFF"],"pattern":1},
		"alignment":{"horizontal":"center", "vertical":"center"},
		"border":[
			{"type":"left","color":"#000000","style":1},
			{"type":"bottom","color":"#FFFFFF","style":1},
			{"type":"right","color":"#FFFFFF","style":1},
			{"type":"top","color":"#FFFFFF","style":1}
		]
	}`,
	"topBorderThinSupervisor": `{
		"font":{"color":"#3333CC"},
		"fill":{"type":"pattern","color":["#99CCFF"],"pattern":1},
		"alignment":{"horizontal":"center", "vertical":"center"},
		"border":[
			{"type":"top","color":"#000000","style":1},
			{"type":"bottom","color":"#FFFFFF","style":1},
			{"type":"left","color":"#FFFFFF","style":1},
			{"type":"right","color":"#FFFFFF","style":1}
		]
	}`,
	"topDotRightContinuousBorderSupervisor": `{
		"font":{"color":"#3333CC"},
		"fill":{"type":"pattern","color":["#99CCFF"],"pattern":1},
		"alignment":{"horizontal":"center", "vertical":"center"},
		"border":[
			{"type":"top","color":"#000000","style":7},
			{"type":"right","color":"#000000","style":1},
			{"type":"left","color":"#FFFFFF","style":1},
			{"type":"bottom","color":"#000000","style":7}
		]
	}`,
	"bottomBoldRightLeftThinBorderSupervisor": `{
		"font":{"color":"#3333CC"},
		"fill":{"type":"pattern","color":["#99CCFF"],"pattern":1},
		"alignment":{"horizontal":"center", "vertical":"center"},
		"border":[
			{"type":"bottom","color":"#000000","style":2},
			{"type":"left","color":"#000000","style":1},
			{"type":"right","color":"#000000","style":1},
			{"type":"top","color":"#FFFFFF","style":1}
		]
	}`,
}

var EvaluationListCategoriesEN = map[string]string{
	"employeeId": "Employee id",
	"name": "Name",
	"quarter": "Quarter",
	"year": "Year",
	"branch": "Branch",
	"point": "Point",
	"rank": "Rank",
}

var EvaluationListCategoriesJP = map[string]string{
	"employeeId": "社員番号",
	"name": "氏名",
	"quarter": "四半",
	"year": "年第",
	"branch": "ブランチ",
	"point": "評価点合計",
	"rank": "評価結果",
}

var EvaluationListCategoriesVN = map[string]string{
	"employeeId": "Mã nhân viên",
	"name": "Họ và tên",
	"quarter": "Qúy",
	"year": "Năm",
	"branch": "Chi nhánh",
	"point": "Điểm",
	"rank": "Xếp hạng",
}
