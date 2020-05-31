package helpers

func GetArticleBody(body string) string {
	return BBCodeCompiler.Compile(body)
}
