package dialect

import "strings"

// OracleDialect implements Oracle-specific features
type OracleDialect struct{}

func (d *OracleDialect) Name() string {
	return "Oracle"
}

func (d *OracleDialect) QuoteIdentifier(identifier string) string {
	return `"` + identifier + `"`
}

func (d *OracleDialect) SupportsFeature(feature Feature) bool {
	switch feature {
	case FeatureCTE:
		return true
	case FeatureWindowFunctions:
		return true
	case FeatureXMLSupport:
		return true
	case FeaturePartitioning:
		return true
	case FeatureFullTextSearch:
		return true // Oracle Text
	case FeatureRecursiveCTE:
		return true
	default:
		return false
	}
}

func (d *OracleDialect) GetKeywords() []string {
	oracle := []string{
		"ROWNUM", "ROWID", "SYSDATE", "SYSTIMESTAMP", "DUAL",
		"CONNECT", "PRIOR", "START", "NOCYCLE", "SIBLINGS",
		"PARTITION", "SUBPARTITION", "COMPRESS", "NOCOMPRESS",
		"LOGGING", "NOLOGGING", "CACHE", "NOCACHE", "PARALLEL", "NOPARALLEL",
		"ENABLE", "DISABLE", "VALIDATE", "NOVALIDATE", "RELY", "NORELY",
		"DEFERRABLE", "INITIALLY", "IMMEDIATE", "DEFERRED",
		"PRESERVE", "PURGE", "FLASHBACK", "VERSIONS", "SCN", "TIMESTAMP",
		"UNPIVOT", "PIVOT", "XMLELEMENT", "XMLATTRIBUTES", "XMLFOREST",
	}
	return append(CommonKeywords, oracle...)
}

func (d *OracleDialect) GetDataTypes() []string {
	oracle := []string{
		"NUMBER", "VARCHAR2", "NVARCHAR2", "CLOB", "NCLOB", "BLOB",
		"LONG", "LONG RAW", "RAW", "ROWID", "UROWID",
		"DATE", "TIMESTAMP", "TIMESTAMP WITH TIME ZONE", "TIMESTAMP WITH LOCAL TIME ZONE",
		"INTERVAL YEAR TO MONTH", "INTERVAL DAY TO SECOND",
		"XMLTYPE", "URITYPE", "BFILE", "BINARY_FLOAT", "BINARY_DOUBLE",
	}
	return append(CommonDataTypes, oracle...)
}

func (d *OracleDialect) IsReservedWord(word string) bool {
	keywords := d.GetKeywords()
	upper := strings.ToUpper(word)
	for _, keyword := range keywords {
		if keyword == upper {
			return true
		}
	}
	return false
}

func (d *OracleDialect) GetLimitSyntax() LimitSyntax {
	return LimitSyntaxOracle
}
