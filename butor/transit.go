package butor

func newTransit() *transit {
	return &transit{subs: make(map[string]map[uint64][]string), subBuckets: make(map[string][]uint64)}
}

// 订阅中转
type transit struct {
	subs       map[string]map[uint64][]string //订阅值快照
	subBuckets map[string][]uint64            //订阅桶全部数据快照
}

func (t *transit) bucketTransit(bucket string, id uint64) {
	t.subBuckets[bucket] = append(t.subBuckets[bucket], id)
}

func (t *transit) subTransit(bucketName string, values []string, id uint64) {
	if _, ok := t.subs[bucketName]; !ok {
		t.subs[bucketName] = make(map[uint64][]string)
	}
	t.subs[bucketName][id] = append(t.subs[bucketName][id], values...)
}
