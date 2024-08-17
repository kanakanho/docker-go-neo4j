#!/bin/bash
set -euC

# EXTENSION_SCRIPTはコンテナが起動するたびにコールされるため、
# import処理が実施済かフラグファイルの有無をチェック
if [ -f /import/done ]; then
    echo "Skip import process"
    exit 0
fi

# データを全削除
echo "Delete database started."
rm -rf /data/databases
rm -rf /data/transactions
echo "Delete database finished."

# CSVデータのインポート
echo "Start the data import process"
neo4j-admin database import full \
  --nodes=/import/actors.csv \
  --relationships=/import/roles.csv
echo "Complete the data import process"

# import処理の完了フラグファイルの作成
echo "Start creating flag file"
touch /import/done
echo "Complete creating flag file"

# EXTENSION_SCRIPTはroot権限で実行されるため、本スクリプトを実行すると、
# /dataと/logsの所有者がrootになってしまうので、所有者をneo4jに変更。
# 所有者を変更しないとNeo4jが起動できません。
echo "Start ownership change"
chown -R neo4j:neo4j /data
chown -R neo4j:neo4j /logs
echo "Complete ownership change"
