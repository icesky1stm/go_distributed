package grades

// 包加载的时候就运行
func init() {
	students = []Student{
		{
			ID:   1,
			Name: "王大",
			Grades: []Grade{
				{
					Title: "Quiz 1",
					Type:  GradeQuiz,
					Score: 85,
				},
				{
					Title: "Final Exam",
					Type:  GradeExam,
					Score: 94,
				},
				{
					Title: "Quiz 2",
					Type:  GradeQuiz,
					Score: 82,
				},
			},
		},
		{
			ID:   2,
			Name: "赵二",
			Grades: []Grade{
				{
					Title: "Quiz 1",
					Type:  GradeQuiz,
					Score: 100,
				},
				{
					Title: "Final Exam",
					Type:  GradeExam,
					Score: 100,
				},
				{
					Title: "Quiz 2",
					Type:  GradeQuiz,
					Score: 81,
				},
			},
		},
		{
			ID:   3,
			Name: "孙三",
			Grades: []Grade{
				{
					Title: "Quiz 1",
					Type:  GradeQuiz,
					Score: 67,
				},
				{
					Title: "Final Exam",
					Type:  GradeExam,
					Score: 0,
				},
				{
					Title: "Quiz 2",
					Type:  GradeQuiz,
					Score: 75,
				},
			},
		},
	}
}
