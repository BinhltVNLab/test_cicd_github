package configs

const (
	TOPCV = 1
	VIETNAMWORKS = 2
	LINKEDIN = 3
	FACEBOOK = 4
	ITVIEC = 5
	JOBNOW = 6
	CAREERBUILDER = 7

	CVPENDING = 1
	CVPASSROUNDONE = 2
	CVPASSFINAL = 3
	CVREJECT = 4
	CVNOTPASS = 5
	CVINTERVIEW = 6
)

var MediasRecruitment = map[int]string{
	TOPCV: "TopCV",
	VIETNAMWORKS: "Vietnamworks",
	LINKEDIN: "LinkedIn",
	FACEBOOK: "Facebook",
	ITVIEC: "ITviec",
	JOBNOW: "JobNow",
	CAREERBUILDER: "CareerBuilder",
}

var CvStatuses = map[int]string{
	CVPENDING: "Pending",
	CVPASSROUNDONE: "Pass round 1",
	CVPASSFINAL: "Pass final",
	CVREJECT: "Reject",
	CVNOTPASS: "Not pass",
	CVINTERVIEW: "Interview appointment",
}
