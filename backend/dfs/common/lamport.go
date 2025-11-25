package common

import "github.com/google/uuid"

// CompareLamport compares two Lamport timestamps using node ID as a tie-breaker.
// Returns:
// -1 if (t1,id1) < (t2,id2)
//
//	0 if (t1,id1) == (t2,id2)
//	1 if (t1,id1) > (t2,id2)
func CompareLamport(t1 int, id1 uuid.UUID, t2 int, id2 uuid.UUID) int {
	if t1 < t2 {
		return -1
	}
	if t1 > t2 {
		return 1
	}
	// Tie-break by node ID to achieve a total order
	s1 := id1.String()
	s2 := id2.String()
	if s1 < s2 {
		return -1
	}
	if s1 > s2 {
		return 1
	}
	return 0
}

func LessLamport(t1 int, id1 uuid.UUID, t2 int, id2 uuid.UUID) bool {
	return CompareLamport(t1, id1, t2, id2) == -1
}

func CompareMessageLamport(MessageWithTime1, MessageWithTime2 MessageWithTime) int {
	return CompareLamport(MessageWithTime1.LogicalTime, MessageWithTime1.Message.From, MessageWithTime2.LogicalTime, MessageWithTime2.Message.From)
}

func CompareLockRequests(a, b *LockRequest) bool {
	if a.LogicalTime != b.LogicalTime {
		return a.LogicalTime < b.LogicalTime
	}
	return a.NodeID.String() < b.NodeID.String()
}
