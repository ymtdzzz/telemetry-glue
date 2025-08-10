package analyze

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ymtdzzz/telemetry-glue/internal/analyzer"
)

// Flags holds flags for analyze command
type Flags struct {
	Type     string
	Provider string
	Model    string
	Language string
}

// AnalyzeCmd creates the analyze subcommand
func AnalyzeCmd() *cobra.Command {
	flags := &Flags{}

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze telemetry data using LLM",
		Long: `Analyze telemetry data from stdin using Large Language Models.
This command reads JSON telemetry data from stdin (typically piped from other commands)
and performs AI-powered analysis based on the specified analysis type.

The command supports combining data from multiple backends by reading multiple JSON objects
from stdin, each on a separate line.

Examples:
  # Analyze performance bottlenecks
  telemetry-glue newrelic spans --trace-id abc123 --format json | \\
    telemetry-glue analyze --type duration --provider vertexai --model gemini

  # Combine multiple data sources for comprehensive analysis
  telemetry-glue newrelic spans --trace-id abc123 --format json | \\
    telemetry-glue gcp logs --project-id my-project --trace-id abc123 --format json | \\
    telemetry-glue analyze --type error --provider vertexai --model gemini`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAnalyze(flags)
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&flags.Type, "type", "t", "", "Analysis type (duration, error) (required)")
	cmd.Flags().StringVarP(&flags.Provider, "provider", "p", "", "LLM provider (vertexai) (required)")
	cmd.Flags().StringVarP(&flags.Model, "model", "m", "", "LLM model name (e.g., gemini) (required)")
	cmd.Flags().StringVarP(&flags.Language, "language", "l", "en", "Output language (en, ja)")

	// Mark required flags
	cmd.MarkFlagRequired("type")
	cmd.MarkFlagRequired("provider")
	cmd.MarkFlagRequired("model")

	return cmd
}

func runAnalyze(flags *Flags) error {
	// Validate input parameters
	if flags.Type != "duration" && flags.Type != "error" {
		return fmt.Errorf("unsupported analysis type: %s (supported: duration, error)", flags.Type)
	}

	if flags.Provider != "vertexai" && flags.Provider != "mock" {
		return fmt.Errorf("unsupported provider: %s (supported: vertexai, mock)", flags.Provider)
	}

	if flags.Language != "en" && flags.Language != "ja" {
		return fmt.Errorf("unsupported language: %s (supported: en, ja)", flags.Language)
	}

	// Read and aggregate data from stdin
	aggregator := analyzer.NewDataAggregator()
	if err := aggregator.ReadFromStdin(os.Stdin); err != nil {
		return fmt.Errorf("failed to read data from stdin: %w", err)
	}

	combined := aggregator.GetCombinedData()

	// Print summary of what we received
	fmt.Printf("Aggregated data: %s\n", combined.Summary())

	// Create LLM provider based on flags
	var provider analyzer.LLMProvider
	var err error

	switch flags.Provider {
	case "vertexai":
		// Use real Vertex AI provider with Application Default Credentials
		provider, err = analyzer.NewVertexAIProvider("o11y-ymtdzzz", "us-central1", flags.Model)
		if err != nil {
			return fmt.Errorf("failed to create Vertex AI provider: %w", err)
		}
		fmt.Println("Using Vertex AI provider with Application Default Credentials")
	case "mock":
		provider = analyzer.NewMockProvider(generateMockAnalysis(flags.Type, flags.Language, combined))
		fmt.Println("Using mock provider for testing")
	default:
		return fmt.Errorf("unsupported provider: %s", flags.Provider)
	}

	// Create analyzer
	llmAnalyzer := analyzer.NewAnalyzer(provider, flags.Provider, flags.Model)
	defer llmAnalyzer.Close()

	// Perform analysis
	ctx := context.Background()
	analysisType := analyzer.AnalysisType(flags.Type)

	result, err := llmAnalyzer.AnalyzeWithLanguage(ctx, analysisType, combined, flags.Language)
	if err != nil {
		return fmt.Errorf("failed to perform analysis: %w", err)
	}

	// Output result
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("# %s Analysis Report\n\n", strings.Title(flags.Type))
	fmt.Printf("**Provider:** %s (%s)\n", result.Provider, result.Model)
	fmt.Printf("**Data Summary:** %s\n\n", result.Summary)
	fmt.Println("## Analysis Results")
	fmt.Println()
	fmt.Println(result.Content)

	return nil
}

// generateMockAnalysis generates a mock analysis for testing
func generateMockAnalysis(analysisType string, language string, data *analyzer.CombinedData) string {
	switch analysisType {
	case "duration":
		if language == "ja" {
			return fmt.Sprintf(`# パフォーマンス分析レポート

## 概要
%s の分析に基づき、以下のパフォーマンス特性が確認されました:

## 主要な発見事項
- **分析対象スパン数**: %d
- **ログエントリー数**: %d  
- **トレース数**: %d

## パフォーマンスボトルネック
1. **最も遅い処理**: データベースクエリが主要なボトルネックとして確認
2. **リソース競合**: サービス間通信において高いレイテンシを観測
3. **メモリ使用量**: 一部のサービスでメモリ消費量の増大パターンを確認

## 推奨事項
1. **データベース最適化**: 頻繁にクエリされるフィールドにインデックスの追加を検討
2. **コネクションプーリング**: 外部サービス呼び出し用のコネクションプーリングを実装
3. **キャッシュ戦略**: 頻繁にアクセスされるデータのキャッシュレイヤーを導入
4. **ロードバランシング**: サービスインスタンス間の負荷分散を見直し

## 次のステップ
- 推奨最適化の実装監視
- 主要パフォーマンス閾値のアラート設定
- 定期的なパフォーマンスレビューのスケジュール

*注意: これはモック分析です。実際の分析では具体的なメトリクスと詳細な洞察が提供されます。*`,
				data.Summary(), len(data.Spans), len(data.Logs), len(data.Traces))
		}
		return fmt.Sprintf(`# Performance Analysis Report

## Executive Summary
Based on the analysis of %s, the following performance characteristics were identified:

## Key Findings
- **Total Spans Analyzed**: %d
- **Total Log Entries**: %d  
- **Total Traces**: %d

## Performance Bottlenecks
1. **Slowest Operations**: Database queries appear to be the primary bottleneck
2. **Resource Contention**: High latency observed in service-to-service communication
3. **Memory Usage**: Some services showing elevated memory consumption patterns

## Recommendations
1. **Database Optimization**: Consider adding indexes for frequently queried fields
2. **Connection Pooling**: Implement connection pooling for external service calls
3. **Caching Strategy**: Introduce caching layer for frequently accessed data
4. **Load Balancing**: Review load distribution across service instances

## Next Steps
- Monitor implementation of recommended optimizations
- Set up alerts for key performance thresholds
- Schedule regular performance reviews

*Note: This is a mock analysis. Real analysis would provide specific metrics and detailed insights.*`,
			data.Summary(), len(data.Spans), len(data.Logs), len(data.Traces))

	case "error":
		if language == "ja" {
			return fmt.Sprintf(`# エラー分析レポート

## 概要
%s のエラーパターン分析により、緊急対応が必要な重要な問題が明らかになりました。

## 主要な発見事項
- **分析対象スパン数**: %d
- **ログエントリー数**: %d
- **トレース数**: %d

## エラーパターン
1. **HTTP 500エラー**: 支払い処理における内部サーバーエラーが頻発
2. **タイムアウト問題**: ピークトラフィック時のサービスタイムアウト
3. **データベース接続エラー**: 断続的なコネクションプール枯渇

## 根本原因分析
1. **主要原因**: 下流サービスにおける不十分なエラーハンドリング
2. **副次原因**: 高負荷時のリソース制約
3. **寄与要因**: サーキットブレーカーパターンの不備

## 影響評価
- **ビジネス影響**: 高 - 顧客取引に影響
- **発生頻度**: リクエストの15%%でエラーが発生
- **復旧時間**: 自動復旧まで平均30秒

## 緊急対応が必要な項目
1. **サーキットブレーカーの実装**: カスケード障害の防止
2. **コネクションプールサイズの増加**: データベース接続問題への対処
3. **監視の強化**: 詳細なエラー追跡とアラートの追加
4. **グレースフル・デグラデーション**: フォールバック機構の実装

## 長期的改善項目
- カオスエンジニアリングテスト
- エラーハンドリングパターンの強化
- 可観測性向上のためのサービスメッシュ実装

*注意: これはモック分析です。実際の分析では具体的なエラーコードとトレースが提供されます。*`,
				data.Summary(), len(data.Spans), len(data.Logs), len(data.Traces))
		}
		return fmt.Sprintf(`# Error Analysis Report

## Executive Summary
Error pattern analysis of %s reveals several critical issues requiring immediate attention.

## Key Findings
- **Total Spans Analyzed**: %d
- **Total Log Entries**: %d
- **Total Traces**: %d

## Error Patterns
1. **HTTP 500 Errors**: Frequent internal server errors in payment processing
2. **Timeout Issues**: Service timeouts during peak traffic periods
3. **Database Connection Errors**: Intermittent connection pool exhaustion

## Root Cause Analysis
1. **Primary Cause**: Insufficient error handling in downstream services
2. **Secondary Cause**: Resource constraints during high load
3. **Contributing Factor**: Lack of circuit breaker patterns

## Impact Assessment
- **Business Impact**: High - affecting customer transactions
- **Frequency**: 15%% of requests experiencing errors
- **Recovery Time**: Average 30 seconds for automatic recovery

## Immediate Actions Required
1. **Implement Circuit Breakers**: Prevent cascade failures
2. **Increase Connection Pool Size**: Address database connectivity issues
3. **Enhanced Monitoring**: Add detailed error tracking and alerting
4. **Graceful Degradation**: Implement fallback mechanisms

## Long-term Improvements
- Chaos engineering testing
- Enhanced error handling patterns
- Service mesh implementation for better observability

*Note: This is a mock analysis. Real analysis would provide specific error codes and traces.*`,
			data.Summary(), len(data.Spans), len(data.Logs), len(data.Traces))

	default:
		return "Mock analysis: Analysis type not supported"
	}
}
