package dialect

import "strings"

// SQLServerDialect implements SQL Server-specific features
type SQLServerDialect struct{}

func (d *SQLServerDialect) Name() string {
	return "SQL Server"
}

func (d *SQLServerDialect) QuoteIdentifier(identifier string) string {
	return "[" + identifier + "]"
}

func (d *SQLServerDialect) SupportsFeature(feature Feature) bool {
	switch feature {
	case FeatureCTE:
		return true
	case FeatureWindowFunctions:
		return true
	case FeatureJSONSupport:
		return true // SQL Server 2016+
	case FeatureXMLSupport:
		return true
	case FeatureUpsert:
		return true // MERGE statement
	default:
		return false
	}
}

func (d *SQLServerDialect) GetKeywords() []string {
	sqlserver := []string{
		"IDENTITY", "UNIQUEIDENTIFIER", "NVARCHAR", "NCHAR", "NTEXT",
		"MONEY", "SMALLMONEY", "DATETIME2", "DATETIMEOFFSET", "SMALLDATETIME",
		"TOP", "OFFSET", "FETCH", "NEXT", "ROWS", "ONLY",
		"MERGE", "USING", "MATCHED", "OUTPUT", "INSERTED", "DELETED",
		"CROSS", "APPLY", "OUTER", "APPLY", "PIVOT", "UNPIVOT",
		"NOLOCK", "READUNCOMMITTED", "READCOMMITTED", "REPEATABLEREAD", "SERIALIZABLE",
		"ROWLOCK", "PAGLOCK", "TABLOCK", "TABLOCKX", "UPDLOCK", "XLOCK",
	}
	return append(CommonKeywords, sqlserver...)
}

func (d *SQLServerDialect) GetDataTypes() []string {
	sqlserver := []string{
		"UNIQUEIDENTIFIER", "MONEY", "SMALLMONEY", "DATETIME2", "DATETIMEOFFSET",
		"SMALLDATETIME", "NVARCHAR", "NCHAR", "NTEXT", "IMAGE",
		"HIERARCHYID", "GEOGRAPHY", "GEOMETRY", "SQL_VARIANT",
	}
	return append(CommonDataTypes, sqlserver...)
}

func (d *SQLServerDialect) IsReservedWord(word string) bool {
	keywords := d.GetKeywords()
	upper := strings.ToUpper(word)
	for _, keyword := range keywords {
		if keyword == upper {
			return true
		}
	}
	return false
}

func (d *SQLServerDialect) GetLimitSyntax() LimitSyntax {
	return LimitSyntaxSQLServer
}
