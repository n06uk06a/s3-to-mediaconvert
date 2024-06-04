# Launch Media Convert Job

AWSのMediaConvertのジョブを起動するfunction

## ビルド
```
$ make build
```

## IAM Role
| 名前 | 権限 |
|-----|-----|
| LambdaMediaConvert | AdministratorAccess |
| MediaConvert | AmazonS3FullAccess, AmazonAPIGatewayInvokeFullAccess |

## Lambdaの設定
* build/s3_to_convert をzip圧縮
* AWSコンソールのLambdaから作成してzipファイルをアップロード

### 設定
| 設定箇所 | 値 |
|-------|----|
| 関数名 | GoMediaConvert |
| コードエントリタイプ | .zipファイルをアップロード |
| ランタイム | Go 1.x |
| ハンドラ | s3_to_mediaconvert |
| 実行ロール | 既存のロールを使用する |
| 既存のロール | LambdaMediaConvert |

### 環境変数
| 変数名 | 説明 |
|-------|-----|
| DEST_BUCKET | 変換後データの保存先バケット名 |
| QUEUE | MediaConvertで変換するジョブが使用するキュー |
| ROLE | MediaConvertに紐づくIAM Role |


## S3の設定
* この設定をすることで変換をするバケットに設定

イベント設定
| 設定箇所 | 値 | 説明 |
|-------|----|------|
| イベント | すべてのオブジェクト作成 ||
| プレフィックス | movie/articles | この下だけでイベント発火 |
| 送信先 | Lambda関数 ||
| Lambda | GoMediaConvert ||
