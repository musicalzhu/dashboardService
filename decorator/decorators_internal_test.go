package decorator

import (
	"strings"
	"testing"
	"time"
)

// 本文件使用 package decorator（内部测试包），
// 可以访问 timedGrouper / loggedGrouper / recoveredGrouper 等未导出的命名函数。
// 这些命名函数是 WithXxx 的内层实现，直接测试它们比通过 WithXxx 间接测试更精准。

// TestTimedGrouper 直接测试 timedGrouper 的计时行为。
func TestTimedGrouper(t *testing.T) {
	var recorded time.Duration
	fake := func(_ []Server) []ClusterGroup { return []ClusterGroup{} }
	g := timedGrouper(fake, func(d time.Duration) { recorded = d })
	g(nil)
	if recorded < 0 {
		t.Errorf("耗时不应为负: %v", recorded)
	}
}

// TestTimedGrouperStatePerCall 验证 start 是每次调用独立的。
//
// 如果把 start 提到 timedGrouper 外部，多次调用会互相干扰。
func TestTimedGrouperStatePerCall(t *testing.T) {
	var durations []time.Duration
	slow := func(_ []Server) []ClusterGroup {
		time.Sleep(time.Millisecond)
		return []ClusterGroup{}
	}
	g := timedGrouper(slow, func(d time.Duration) { durations = append(durations, d) })
	g(nil)
	g(nil)
	if len(durations) != 2 {
		t.Fatalf("应记录 2 次，实际 %d", len(durations))
	}
	for i, d := range durations {
		if d < time.Millisecond {
			t.Errorf("第 %d 次耗时 %v 应 >= 1ms", i, d)
		}
	}
}

// TestLoggedGrouperStartDonePair 验证日志成对输出。
func TestLoggedGrouperStartDonePair(t *testing.T) {
	var logs []string
	fake := func(_ []Server) []ClusterGroup { return []ClusterGroup{{}} }
	g := loggedGrouper(fake, func(format string, _ ...any) { logs = append(logs, format) })
	g(nil)
	if len(logs) != 2 {
		t.Fatalf("应 2 条日志，实际 %d", len(logs))
	}
	if !strings.Contains(logs[0], "start") || !strings.Contains(logs[1], "done") {
		t.Errorf("日志应成 start/done 对: %v", logs)
	}
}

// TestRecoveredGrouperNonNilOnPanic 验证 panic 时返回非 nil 空切片。
func TestRecoveredGrouperNonNilOnPanic(t *testing.T) {
	panicky := func(_ []Server) []ClusterGroup { panic("boom") }
	var captured any
	g := recoveredGrouper(panicky, func(v any) { captured = v })
	result := g(nil)
	if result == nil {
		t.Error("panic 后应返回非 nil 空切片")
	}
	if len(result) != 0 {
		t.Errorf("panic 后应返回空切片，实际 %d", len(result))
	}
	if captured != "boom" {
		t.Errorf("panic 值 = %v, want boom", captured)
	}
}

// TestRecoveredGrouperNoPanic 验证未 panic 时透传结果。
func TestRecoveredGrouperNoPanic(t *testing.T) {
	want := []ClusterGroup{{}}
	g := recoveredGrouper(func(_ []Server) []ClusterGroup { return want }, nil)
	got := g(nil)
	if len(got) != len(want) {
		t.Error("未 panic 时应透传结果")
	}
}
