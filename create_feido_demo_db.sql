CREATE TABLE "revoked_eids" (
    "id"	INTEGER,
    "doc_num" BLOB NOT NULL,
    "doc_country"	BLOB NOT NULL,
    "doc_type"	BLOB NOT NULL,
    UNIQUE(doc_num,doc_country, doc_type),
    PRIMARY KEY("id")
);

INSERT INTO revoked_eids
    VALUES(1,
        x'0c0a0d0a0b0e09030a',
        x'00000D',
        x'0009');

INSERT INTO revoked_eids
    VALUES(2,
       x'112233445566778899',
       x'112233',
       x'1122');