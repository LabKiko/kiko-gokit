/**
* Author: JeffreyBool
* Date: 2021/11/10
* Time: 22:08
* Software: GoLand
 */

package kafka

var _reader Reader

func InitReader(brokers []string, topic, group string) (Reader, error) {
	var err error
	_reader, err = NewReader(brokers, topic, group, ReaderStartOffset(OffsetNewest))
	return _reader, err
}
