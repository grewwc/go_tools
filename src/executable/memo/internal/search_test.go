package internal

import (
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestSearchRecordsMatchesExactMixedToken(t *testing.T) {
	disableRemoteSearchProviders(t)
	cleanup := setupSQLiteTest(t)
	defer cleanup()

	base := time.Unix(1_700_100_000, 0)
	records := []*Record{
		{
			ID:           primitive.NewObjectID(),
			Tags:         []string{"memo"},
			AddDate:      base,
			ModifiedDate: base,
			MyProblem:    true,
			Title:        "xxx链接：https://xxx.example.com",
		},
		{
			ID:           primitive.NewObjectID(),
			Tags:         []string{"memo"},
			AddDate:      base.Add(time.Second),
			ModifiedDate: base.Add(time.Second),
			MyProblem:    true,
			Title:        "xxx文档：用于排查问题",
		},
	}
	for _, record := range records {
		record.Save(true)
	}

	results := SearchRecords("xxx链接", 5, true, nil, false, false)
	if len(results) == 0 {
		t.Fatal("expected fuzzy search results")
	}
	if results[0].Record.ID != records[0].ID {
		t.Fatalf("expected link record first, got %+v", results[0].Record)
	}
}

func TestSearchRecordsRanksFuzzyLinkIntent(t *testing.T) {
	disableRemoteSearchProviders(t)
	cleanup := setupSQLiteTest(t)
	defer cleanup()

	base := time.Unix(1_700_200_000, 0)
	best := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"infra"},
		AddDate:      base,
		ModifiedDate: base,
		MyProblem:    true,
		Title:        "生产数据库链接：https://db.example.com",
	}
	other := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"infra"},
		AddDate:      base.Add(time.Second),
		ModifiedDate: base.Add(time.Second),
		MyProblem:    true,
		Title:        "生产数据库账号：admin",
	}
	best.Save(true)
	other.Save(true)

	results := SearchRecords("数据库地址", 5, true, nil, false, false)
	if len(results) == 0 {
		t.Fatal("expected fuzzy search results")
	}
	if results[0].Record.ID != best.ID {
		t.Fatalf("expected database link record first, got %+v", results[0].Record)
	}
	if len(results) > 1 && results[1].Score >= results[0].Score {
		t.Fatalf("expected first result to have highest score: %+v", results)
	}
}

func TestSearchRecordsCombinesTitleAndTags(t *testing.T) {
	disableRemoteSearchProviders(t)
	cleanup := setupSQLiteTest(t)
	defer cleanup()

	base := time.Unix(1_700_300_000, 0)
	best := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"交接文档", "team"},
		AddDate:      base,
		ModifiedDate: base,
		MyProblem:    true,
		Title:        "线上地址：https://handoff.example.com",
	}
	other := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"team"},
		AddDate:      base.Add(time.Second),
		ModifiedDate: base.Add(time.Second),
		MyProblem:    true,
		Title:        "交接文档模板",
	}
	best.Save(true)
	other.Save(true)

	results := SearchRecords("交接文档链接", 10, true, nil, false, false)
	if len(results) == 0 {
		t.Fatal("expected fuzzy search results")
	}
	if results[0].Record.ID != best.ID {
		t.Fatalf("expected title+tag match first, got %+v", results[0].Record)
	}
}

func TestSearchRecordsDoesNotApplyHardThreeResultCap(t *testing.T) {
	disableRemoteSearchProviders(t)
	cleanup := setupSQLiteTest(t)
	defer cleanup()

	base := time.Unix(1_700_400_000, 0)
	for i := 0; i < 4; i++ {
		record := &Record{
			ID:           primitive.NewObjectID(),
			Tags:         []string{"foo"},
			AddDate:      base.Add(time.Duration(i) * time.Second),
			ModifiedDate: base.Add(time.Duration(i) * time.Second),
			MyProblem:    true,
			Title:        "foo链接：https://example.com/" + string(rune('a'+i)),
		}
		record.Save(true)
	}

	results := SearchRecords("foo链接", 100, true, nil, false, false)
	if len(results) != 4 {
		t.Fatalf("expected all 4 results without a hard cap, got %d", len(results))
	}

	results = SearchRecords("foo链接", 2, true, nil, false, false)
	if len(results) != 2 {
		t.Fatalf("expected explicit smaller limit to be respected, got %d", len(results))
	}
}

func TestSearchRecordsMatchesRelatedChineseIntent(t *testing.T) {
	disableRemoteSearchProviders(t)
	cleanup := setupSQLiteTest(t)
	defer cleanup()

	base := time.Unix(1_700_500_000, 0)
	best := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"handoff", "交接文档"},
		AddDate:      base,
		ModifiedDate: base,
		MyProblem:    true,
		Title:        "交接文档链接：https://handoff.example.com/doc",
	}
	other := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"handoff", "流程"},
		AddDate:      base.Add(time.Second),
		ModifiedDate: base.Add(time.Second),
		MyProblem:    true,
		Title:        "交接流程说明",
	}
	best.Save(true)
	other.Save(true)

	results := SearchRecords("交接相关", 10, true, nil, false, false)
	if len(results) == 0 {
		t.Fatal("expected related chinese intent query to return results")
	}
	if results[0].Record.ID != best.ID {
		t.Fatalf("expected link record first for related query, got %+v", results[0].Record)
	}
}

func TestSearchRecordsExcludeFinishedByDefault(t *testing.T) {
	disableRemoteSearchProviders(t)
	cleanup := setupSQLiteTest(t)
	defer cleanup()

	base := time.Unix(1_700_550_000, 0)
	active := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"links"},
		AddDate:      base,
		ModifiedDate: base,
		MyProblem:    true,
		Title:        "handoff 当前文档：https://current.example.com/doc",
	}
	finished := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"links"},
		AddDate:      base.Add(time.Second),
		ModifiedDate: base.Add(time.Second),
		MyProblem:    true,
		Finished:     true,
		Title:        "handoff 历史文档：https://archive.example.com/doc",
	}
	active.Save(true)
	finished.Save(true)

	results := SearchRecords("handoff", 10, false, nil, false, false)
	if len(results) == 0 {
		t.Fatal("expected active search results")
	}
	for _, result := range results {
		if result.Record.ID == finished.ID {
			t.Fatalf("expected finished records to be excluded by default: %+v", results)
		}
	}

	results = SearchRecords("handoff", 10, true, nil, false, false)
	foundFinished := false
	for _, result := range results {
		if result.Record.ID == finished.ID {
			foundFinished = true
			break
		}
	}
	if !foundFinished {
		t.Fatalf("expected finished records to be included when includeFinished=true: %+v", results)
	}
}

func TestSearchRecordsRejectsGenericDatasetLinkNoiseForDocLinkQuery(t *testing.T) {
	disableRemoteSearchProviders(t)
	cleanup := setupSQLiteTest(t)
	defer cleanup()

	base := time.Unix(1_700_575_000, 0)
	best := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"links"},
		AddDate:      base,
		ModifiedDate: base,
		MyProblem:    true,
		Title:        "风神交接文档：https://handoff.example.com/doc",
	}
	partiallyRelevant := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"json_固化"},
		AddDate:      base.Add(time.Second),
		ModifiedDate: base.Add(time.Second),
		MyProblem:    true,
		Title:        "json固化相关文档\n风神性能交接文档：url：https://perf.example.com/doc",
	}
	noise := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"transfer"},
		AddDate:      base.Add(2 * time.Second),
		ModifiedDate: base.Add(2 * time.Second),
		MyProblem:    true,
		Title:        "fabric 数据集链接：https://data.example.com/query\n子数据集链接：https://data.example.com/subquery",
	}
	best.Save(true)
	partiallyRelevant.Save(true)
	noise.Save(true)

	results := SearchRecords("交接文档链接", 10, false, nil, false, false)
	if len(results) == 0 {
		t.Fatal("expected doc-link search results")
	}
	if results[0].Record.ID != best.ID && results[0].Record.ID != partiallyRelevant.ID {
		t.Fatalf("expected a交接文档 result first, got %+v", results[0].Record)
	}
	for _, result := range results {
		if result.Record.ID == noise.ID {
			t.Fatalf("expected generic dataset link noise to be excluded, got %+v", results)
		}
	}
}

func TestSearchRecordsPreferLinksRecordWithMultipleMatches(t *testing.T) {
	disableRemoteSearchProviders(t)
	cleanup := setupSQLiteTest(t)
	defer cleanup()

	base := time.Unix(1_700_600_000, 0)
	dbRecord := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"db"},
		AddDate:      base,
		ModifiedDate: base,
		MyProblem:    true,
		Title:        "dataagent_db:\nurl: https://db.example.com",
	}
	todoRecord := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"todo"},
		AddDate:      base.Add(time.Second),
		ModifiedDate: base.Add(time.Second),
		MyProblem:    true,
		Title:        "1. dataagent_be 代码run",
	}
	mainRecord := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"main"},
		AddDate:      base.Add(2 * time.Second),
		ModifiedDate: base.Add(2 * time.Second),
		MyProblem:    true,
		Title:        "doas -V 0 -p data.agent.llm python -c \"print('ok')\"",
	}
	linksRecord := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"links"},
		AddDate:      base.Add(3 * time.Second),
		ModifiedDate: base.Add(3 * time.Second),
		MyProblem:    true,
		Title:        "28. DataAgent研发找人地图: https://a.example.com\n30. DataAgent调用明细: https://b.example.com\n31. DataAgent查看召回结果: https://c.example.com",
	}
	for _, record := range []*Record{dbRecord, todoRecord, mainRecord, linksRecord} {
		record.Save(true)
	}

	results := SearchRecords("dataagent", 10, true, nil, false, false)
	foundLinks := false
	for _, result := range results {
		if result.Record.ID == linksRecord.ID {
			foundLinks = true
			break
		}
	}
	if !foundLinks {
		t.Fatalf("expected links record to be included in search results: %+v", results)
	}
}

func TestSearchPreviewKeepsRelevantLinesOnly(t *testing.T) {
	disableRemoteSearchProviders(t)
	record := &Record{
		ID:    primitive.NewObjectID(),
		Tags:  []string{"json_固化"},
		Title: "json固化相关文档\n风神 JSON 固化：url：https://a.example.com\n风神性能交接文档：url：https://b.example.com\n存量json string自动改写：url：https://c.example.com",
	}

	preview := SearchPreview(record, "交接相关")
	if len(preview) == 0 {
		t.Fatal("expected non-empty search preview")
	}
	joined := strings.Join(preview, "\n")
	if !strings.Contains(joined, "交接文档") {
		t.Fatalf("expected preview to keep the relevant line, got %q", joined)
	}
	if strings.Contains(joined, "json固化相关文档") {
		t.Fatalf("expected preview to omit unrelated heading, got %q", joined)
	}
}

func TestSearchPreviewReturnsAllRelevantLinksLines(t *testing.T) {
	disableRemoteSearchProviders(t)
	record := &Record{
		ID:   primitive.NewObjectID(),
		Tags: []string{"links"},
		Title: "28. DataAgent研发找人地图: https://a.example.com\n" +
			"30. DataAgent调用明细: https://b.example.com\n" +
			"31. DataAgent查看召回结果: https://c.example.com",
	}

	preview := SearchPreview(record, "dataagent")
	if len(preview) != 3 {
		t.Fatalf("expected all relevant links lines in preview, got %d: %+v", len(preview), preview)
	}
}

func TestSearchPreviewIncludesDetailLineAfterHeading(t *testing.T) {
	disableRemoteSearchProviders(t)
	record := &Record{
		ID:   primitive.NewObjectID(),
		Tags: []string{"db"},
		Title: "dataagent_db:\n" +
			"url: https://db.example.com/workbench",
	}

	preview := SearchPreview(record, "dataagent")
	joined := strings.Join(preview, "\n")
	if !strings.Contains(joined, "dataagent_db") || !strings.Contains(joined, "https://db.example.com/workbench") {
		t.Fatalf("expected heading preview to include the following detail line, got %q", joined)
	}
}

func TestSearchPreviewKeepsFullLongURL(t *testing.T) {
	disableRemoteSearchProviders(t)
	longURL := "https://example.com/very/long/path/with/query?token=abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789"
	record := &Record{
		ID:   primitive.NewObjectID(),
		Tags: []string{"links"},
		Title: "handoff 文档:\n" +
			longURL,
	}

	preview := SearchPreview(record, "handoff")
	joined := strings.Join(preview, "\n")
	if !strings.Contains(joined, longURL) {
		t.Fatalf("expected preview to keep full long url, got %q", joined)
	}
	if strings.Contains(joined, "...") {
		t.Fatalf("expected preview to avoid ellipsis for long urls, got %q", joined)
	}
}

func TestSearchRecordsAvoidsGenericDataNoiseForDataagent(t *testing.T) {
	disableRemoteSearchProviders(t)
	cleanup := setupSQLiteTest(t)
	defer cleanup()

	base := time.Unix(1_700_700_000, 0)
	noise := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"transfer"},
		AddDate:      base,
		ModifiedDate: base,
		MyProblem:    true,
		Title:        "task_id: aeolus_data_db_foo\ntask_id: aeolus_data_table_bar",
	}
	useful := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"links"},
		AddDate:      base.Add(time.Second),
		ModifiedDate: base.Add(time.Second),
		MyProblem:    true,
		Title:        "DataAgent调用明细: https://b.example.com\nDataAgent查看召回结果: https://c.example.com",
	}
	noise.Save(true)
	useful.Save(true)

	results := SearchRecords("dataagent", 10, true, nil, false, false)
	if len(results) == 0 {
		t.Fatal("expected search results")
	}
	if results[0].Record.ID != useful.ID {
		t.Fatalf("expected exact dataagent result to outrank generic data noise, got %+v", results[0].Record)
	}
}

func TestSearchRecordsRequiresStrongShortASCIIQueryMatch(t *testing.T) {
	disableRemoteSearchProviders(t)
	cleanup := setupSQLiteTest(t)
	defer cleanup()

	base := time.Unix(1_700_800_000, 0)
	useful := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"db"},
		AddDate:      base,
		ModifiedDate: base,
		MyProblem:    true,
		Title:        "dataagent_db:\nurl: https://db.example.com/workbench",
	}
	noise := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"dash"},
		AddDate:      base.Add(time.Second),
		ModifiedDate: base.Add(time.Second),
		MyProblem:    true,
		Title:        "glue->be各接口耗时：https://data.bytedance.net/aeolus/pages/dataQuery?appId=2933&dashboardId=1493626",
	}
	useful.Save(true)
	noise.Save(true)

	results := SearchRecords("db链接", 10, true, nil, false, false)
	if len(results) == 0 {
		t.Fatal("expected db search results")
	}
	if results[0].Record.ID != useful.ID {
		t.Fatalf("expected db record to rank first for db链接, got %+v", results[0].Record)
	}
	for _, result := range results {
		if result.Record.ID == noise.ID {
			t.Fatalf("expected dash/url noise to be excluded from db链接 results: %+v", results)
		}
	}
}

func TestSearchRecordsRejectsTransferDatasetNoiseForDBLinkQuery(t *testing.T) {
	disableRemoteSearchProviders(t)
	cleanup := setupSQLiteTest(t)
	defer cleanup()

	base := time.Unix(1_700_810_000, 0)
	useful := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"连接线上db"},
		AddDate:      base,
		ModifiedDate: base,
		MyProblem:    true,
		Title: "dev环境访问线上数据库，需要在common_settings.py 中添加如下内容：\n" +
			"DB_HOST = \"10.226.15.175\"\n" +
			"DB_PORT = 3306\n" +
			"DB_USER = \"aeo7707498154_w\"\n" +
			"DB_PASSWD = \"secret\"\n" +
			"DB_NAME = \"aeolus_db\"",
	}
	noise := &Record{
		ID:           primitive.NewObjectID(),
		Tags:         []string{"transfer"},
		AddDate:      base.Add(time.Second),
		ModifiedDate: base.Add(time.Second),
		MyProblem:    true,
		Title: "fabric 数据集链接：https://data.example.com/query?appId=555159&sid=2220553\n" +
			"子数据集链接：https://data.example.com/subquery?appId=555159&sid=5095973\n" +
			"task_id: t#1769001418189510496#aeolus_theta.aeolus_data_db_aeolus_theta_202601.aeolus_data_table_2_1736174_migrate_v5_prod",
	}
	useful.Save(true)
	noise.Save(true)

	results := SearchRecords("db链接", 10, true, nil, false, false)
	if len(results) == 0 {
		t.Fatal("expected db-link search results")
	}
	if results[0].Record.ID != useful.ID {
		t.Fatalf("expected db access record to rank first for db链接, got %+v", results[0].Record)
	}
	for _, result := range results {
		if result.Record.ID == noise.ID {
			t.Fatalf("expected transfer dataset noise to be excluded from db链接 results: %+v", results)
		}
	}
}
