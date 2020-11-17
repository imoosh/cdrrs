package kafka

import "testing"

func TestConsumer(t *testing.T) {
    for i := 0; i < 4; i++ {
        go Consumer(i)
    }
    select {}
}
