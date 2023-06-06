/**
* Author: JeffreyBool
* Date: 2021/11/10
* Time: 01:27
* Software: GoLand
 */

package logger

func InitDefaultLogger() Logger {
	DefaultLogger = New(
		WithBasePath("../logs"),
		WithLevel(InfoLevel),
		WithConsole(true),
		WithFields(map[string]interface{}{
			"app_id":      "mt",
			"instance_id": "JeffreyBool",
		}),
	)
	return DefaultLogger
}
