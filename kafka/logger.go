/**
* Author: JeffreyBool
* Date: 2021/7/13
* Time: 17:35
* Software: GoLand
 */

package kafka

import (
	"github.com/LabKiko/kiko-gokit/logger"
	"github.com/Shopify/sarama"
)

func init() {
	// https://github.com/Shopify/sarama/issues/959
	sarama.MaxRequestSize = 1000000
	stdLog := logger.DefaultLogger.StdLog()
	sarama.Logger = stdLog
	sarama.DebugLogger = stdLog
}

func InitLogger(log logger.Logger) {
	// stdLog := log.StdLog()
	// sarama.Logger = stdLog
	// sarama.DebugLogger = stdLog
}
