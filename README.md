## WeChat Backup DB Tools

decrypt Backup.db to raw and resource file,manual manager wechat messages.
about `WeChatConnectionServerKey` in remote server. could be found in `-[WXGBackupMgr getConnectionInfo]` Object `GetConnectInfoResponse.key.buffer`

## Build

```bash
$: ./genprotobuf.sh && go build
```

## Install

```bash
$: go install github.com/anonymous5l/wcdb
```

## Decrypt Database

```bash
$: wcdb dump -i <Backup.db> -p <WeChatConnectionServerKey> --output <DecryptBackupDBPath>
```

## Backup Sessions

```bash
$: wcdb session -d <DecryptBackupDBPath>
```

## Chat Message

```bash
$: wcdb chat -m <WithMediaFile> -d <DecryptBackupDBPath> -r <WeChatBackupDirectory> -t <Talker take from session subcommand> -p <WeChatConnectionServerKey>
```