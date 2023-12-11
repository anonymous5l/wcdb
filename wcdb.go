package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"database/sql"
	"errors"
	"golang.org/x/crypto/pbkdf2"
	"hash"
	"io"
)

const (
	HMACSaltMask = 0x3a
	FastKdfIter  = 2
	BlockSize    = 16
	IvSize       = BlockSize
	SaltSize     = BlockSize
)

var (
	SQLiteHead = []byte("SQLite format 3\x00")
)

type SqlCipher struct {
	pass      []byte
	key       []byte
	page      []byte
	pageNum   int
	pageCount int
	pageSize  int
	kdfIter   int
	kdfSalt   []byte
	hmac      hash.Hash
	hmacSize  int
	hmacKey   []byte
	size      int64
	block     cipher.Block
	reader    io.Reader
	reserved  int
	hash      func() hash.Hash
}

func NewSqlcipher(pass []byte, pageSize int, kdfIter int, hash func() hash.Hash, r io.ReadSeeker) (b *SqlCipher, err error) {
	block := &SqlCipher{
		pass:     pass,
		pageSize: pageSize,
		kdfIter:  kdfIter,
		hash:     hash,
		reader:   r,
	}
	if block.size, err = r.Seek(0, io.SeekEnd); err != nil {
		return
	}
	block.pageCount = int(block.size) / pageSize
	if _, err = r.Seek(0, io.SeekStart); err != nil {
		return
	}
	block.kdfSalt = make([]byte, BlockSize, BlockSize)
	if _, err = r.Read(block.kdfSalt); err != nil {
		return
	}
	block.key = pbkdf2.Key(pass, block.kdfSalt, kdfIter, len(pass), hash)
	hmacKdfSalt := make([]byte, BlockSize, BlockSize)
	copy(hmacKdfSalt, block.kdfSalt)
	for i := 0; i < len(hmacKdfSalt); i++ {
		hmacKdfSalt[i] ^= HMACSaltMask
	}
	block.hmacKey = pbkdf2.Key(block.key, hmacKdfSalt, FastKdfIter, len(pass), block.hash)
	block.hmac = hmac.New(hash, block.hmacKey)
	block.hmacSize = block.hmac.Size()

	if block.block, err = aes.NewCipher(block.key); err != nil {
		return
	}
	block.page = make([]byte, pageSize, pageSize)

	reserved := BlockSize + block.hmacSize
	if reserved%BlockSize == 0 {
		block.reserved = reserved
	} else {
		block.reserved = ((reserved / BlockSize) + 1) * BlockSize
	}
	b = block
	return
}

func (w *SqlCipher) PageSize() int {
	return w.pageSize
}

func (w *SqlCipher) pageHmac(pageSize, p int) []byte {
	w.hmac.Reset()
	w.hmac.Write(w.page[:pageSize-w.reserved+IvSize])
	pageNo := make([]byte, 4, 4)
	pageNo[0] = byte(p & 0xff)
	pageNo[1] = byte((p >> 8) & 0xff)
	pageNo[2] = byte((p >> 16) & 0xff)
	pageNo[3] = byte((p >> 24) & 0xff)
	w.hmac.Write(pageNo)
	return w.hmac.Sum(nil)
}

func (w *SqlCipher) Next() bool {
	w.pageNum++
	return w.pageNum <= w.pageCount
}

func (w *SqlCipher) Data() (b []byte, err error) {
	if w.pageNum == 0 {
		return nil, nil
	}

	if w.pageNum > w.pageCount {
		return nil, nil
	}

	var (
		pageSize int
		n        int
	)

	if w.pageNum == 1 {
		pageSize = w.pageSize - SaltSize
	} else {
		pageSize = w.pageSize
	}

	if n, err = w.reader.Read(w.page[:pageSize]); err != nil {
		return
	} else {
		pageSize = n
	}

	pageIv := w.page[pageSize-w.reserved : pageSize-w.reserved+IvSize]
	pageHmac := w.page[pageSize-w.reserved+IvSize : pageSize-w.reserved+IvSize+w.hmacSize]

	if bytes.Compare(w.pageHmac(pageSize, w.pageNum), pageHmac) != 0 {
		return nil, errors.New("invalid hmac hash")
	}

	b = make([]byte, pageSize, pageSize)
	decrypter := cipher.NewCBCDecrypter(w.block, pageIv)
	decrypter.CryptBlocks(b, w.page[:pageSize-w.reserved])
	// meaningless or zero pad just fill pageSize
	// rand.Read(b[pageSize:])
	return
}

type BackupDB struct {
	db *sql.DB
}

func NewBackupDB(filename string) (*BackupDB, error) {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}

	b := &BackupDB{
		db: db,
	}

	return b, nil
}

func (db *BackupDB) Sessions() ([]Session, error) {
	var sessions []Session
	rows, err := db.db.Query("SELECT * FROM Session ORDER BY TotalSize DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var session Session
		if err = rows.Scan(&session.Talker, &session.EndTime, &session.TotalSize,
			&session.NickName, &session.Reserved0, &session.Reserved1,
			&session.Reserved2, &session.Reserved3, &session.StartTime,
			&session.Reserved5); err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func (db *BackupDB) Name2ID() ([]Name2ID, error) {
	var sessions []Name2ID
	rows, err := db.db.Query("SELECT * FROM Name2ID")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var session Name2ID
		if err = rows.Scan(&session.UsrName); err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func (db *BackupDB) MsgSegment(talkerId int) ([]MsgSegment, error) {
	row, err := db.db.Query("SELECT * FROM MsgSegments WHERE talkerId = ?", talkerId)
	if err != nil {
		return nil, err
	}
	defer row.Close()
	var msgs []MsgSegment
	for row.Next() {
		var msg MsgSegment
		if err := row.Scan(&msg.TalkerId, &msg.StartTime, &msg.EndTime,
			&msg.OffSet, &msg.Length, &msg.UsrName, &msg.Status,
			&msg.Reserved1, &msg.FilePath, &msg.SegmentId, &msg.Reserved2,
			&msg.Reserved3); err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}
	return msgs, nil
}

func (db *BackupDB) FileSegment(id int) ([]MsgFileSegment, error) {
	var files []MsgFileSegment
	rows, err := db.db.Query("SELECT * FROM MsgFileSegment WHERE MapKey = ? ORDER BY InnerOffset", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var f MsgFileSegment
		if err = rows.Scan(&f.MapKey, &f.InnerOffSet, &f.Length,
			&f.TotalLen, &f.OffSet, &f.Reserved1, &f.FileName,
			&f.Reserved2, &f.Reserved3, &f.Reserved4); err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, nil
}

func (db *BackupDB) FileSegments() ([]MsgFileSegment, error) {
	var files []MsgFileSegment
	rows, err := db.db.Query("SELECT * FROM MsgFileSegment ORDER BY MapKey, InnerOffset")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var f MsgFileSegment
		if err = rows.Scan(&f.MapKey, &f.InnerOffSet, &f.Length,
			&f.TotalLen, &f.OffSet, &f.Reserved1, &f.FileName,
			&f.Reserved2, &f.Reserved3, &f.Reserved4); err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, nil
}

func (db *BackupDB) Close() error {
	return db.db.Close()
}

func (db *BackupDB) MsgMedia(idStr string) (*MsgMedia, error) {
	var msg MsgMedia
	row := db.db.QueryRow("SELECT * FROM MsgMedia WHERE MediaIdStr = ?", idStr)
	if row.Err() != nil {
		return nil, row.Err()
	}
	if err := row.Scan(&msg.TalkerId, &msg.MediaId, &msg.MsgSegmentId,
		&msg.SrvId, &msg.MD5, &msg.Talker, &msg.MediaIdStr,
		&msg.Reserved0, &msg.Reserved1, &msg.Reserved2); err != nil {
		return nil, err
	}
	return &msg, nil
}
