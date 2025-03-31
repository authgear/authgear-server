package main

import (
	"database/sql/driver"
	"flag"
	"io"
	"os"

	"github.com/iawaknahc/gogenwrapper/pkg/gogenwrapper"
)

func main() {
	var packageName string
	flag.StringVar(&packageName, "package", "", "")

	var variant string
	flag.StringVar(&variant, "variant", "", "")
	flag.Parse()

	switch variant {
	case "Driver":
		g := &gogenwrapper.T{
			PackageName:            packageName,
			StructNameWrapper:      "DriverWrapper",
			StructNameInterceptor:  "DriverInterceptor",
			FunctionNameWrap:       "WrapDriver",
			FunctionTypeNamePrefix: "Driver_",
			InterfaceNameUnwrap:    "DriverUnwrapper",
			FunctionNameUnwrap:     "UnwrapDriver",
			BaseInterface:          new(driver.Driver),
			AdditionalInterfaces: []any{
				new(driver.DriverContext),
			},
		}

		mainFileContents, err := g.GenerateGoSourceFile()
		if err != nil {
			panic(err)
		}

		err = os.WriteFile("driver.go", []byte(mainFileContents), 0o600)
		if err != nil {
			panic(err)
		}

		testFileContents, err := g.GenerateGoTestFile()
		if err != nil {
			panic(err)
		}
		err = os.WriteFile("driver_test.go", []byte(testFileContents), 0o600)
		if err != nil {
			panic(err)
		}
	case "Connector":
		g := &gogenwrapper.T{
			PackageName:            packageName,
			StructNameWrapper:      "ConnectorWrapper",
			StructNameInterceptor:  "ConnectorInterceptor",
			InterfaceNameUnwrap:    "ConnectorUnwrapper",
			FunctionNameWrap:       "WrapConnector",
			FunctionNameUnwrap:     "UnwrapConnector",
			FunctionTypeNamePrefix: "Connector_",
			BaseInterface:          new(driver.Connector),
			AdditionalInterfaces: []any{
				new(io.Closer),
			},
		}

		mainFileContents, err := g.GenerateGoSourceFile()
		if err != nil {
			panic(err)
		}

		err = os.WriteFile("connector.go", []byte(mainFileContents), 0o600)
		if err != nil {
			panic(err)
		}

		testFileContents, err := g.GenerateGoTestFile()
		if err != nil {
			panic(err)
		}
		err = os.WriteFile("connecetor_test.go", []byte(testFileContents), 0o600)
		if err != nil {
			panic(err)
		}
	case "Conn":
		g := &gogenwrapper.T{
			PackageName:            packageName,
			StructNameWrapper:      "ConnWrapper",
			StructNameInterceptor:  "ConnInterceptor",
			FunctionNameWrap:       "WrapConn",
			FunctionTypeNamePrefix: "Conn_",
			InterfaceNameUnwrap:    "ConnUnwrapper",
			FunctionNameUnwrap:     "UnwrapConn",
			BaseInterface:          new(driver.Conn),
			AdditionalInterfaces: []any{
				new(driver.ConnBeginTx),
				new(driver.ConnPrepareContext),
				new(driver.Execer),
				new(driver.ExecerContext),
				new(driver.NamedValueChecker),
				new(driver.Pinger),
				new(driver.Queryer),
				new(driver.QueryerContext),
				new(driver.SessionResetter),
				new(driver.Validator),
			},
		}

		mainFileContents, err := g.GenerateGoSourceFile()
		if err != nil {
			panic(err)
		}

		err = os.WriteFile("conn.go", []byte(mainFileContents), 0o600)
		if err != nil {
			panic(err)
		}

		testFileContents, err := g.GenerateGoTestFile()
		if err != nil {
			panic(err)
		}
		err = os.WriteFile("conn_test.go", []byte(testFileContents), 0o600)
		if err != nil {
			panic(err)
		}
	case "Stmt":
		g := &gogenwrapper.T{
			PackageName:            packageName,
			StructNameWrapper:      "StmtWrapper",
			StructNameInterceptor:  "StmtInterceptor",
			FunctionNameWrap:       "WrapStmt",
			FunctionTypeNamePrefix: "Stmt_",
			InterfaceNameUnwrap:    "StmtUnwrapper",
			FunctionNameUnwrap:     "UnwrapStmt",
			BaseInterface:          new(driver.Stmt),
			AdditionalInterfaces: []any{
				new(driver.ColumnConverter),
				new(driver.NamedValueChecker),
				new(driver.StmtExecContext),
				new(driver.StmtQueryContext),
			},
		}

		mainFileContents, err := g.GenerateGoSourceFile()
		if err != nil {
			panic(err)
		}

		err = os.WriteFile("stmt.go", []byte(mainFileContents), 0o600)
		if err != nil {
			panic(err)
		}

		testFileContents, err := g.GenerateGoTestFile()
		if err != nil {
			panic(err)
		}
		err = os.WriteFile("stmt_test.go", []byte(testFileContents), 0o600)
		if err != nil {
			panic(err)
		}
	case "Rows":
		g := &gogenwrapper.T{
			PackageName:            packageName,
			StructNameWrapper:      "RowsWrapper",
			StructNameInterceptor:  "RowsInterceptor",
			FunctionNameWrap:       "WrapRows",
			FunctionTypeNamePrefix: "Rows_",
			InterfaceNameUnwrap:    "RowsUnwrapper",
			FunctionNameUnwrap:     "UnwrapRows",
			BaseInterface:          new(driver.Rows),
			AdditionalInterfaces: []any{
				new(driver.RowsColumnTypeDatabaseTypeName),
				new(driver.RowsColumnTypeLength),
				new(driver.RowsColumnTypeNullable),
				new(driver.RowsColumnTypePrecisionScale),
				new(driver.RowsColumnTypeScanType),
				new(driver.RowsNextResultSet),
			},
		}

		mainFileContents, err := g.GenerateGoSourceFile()
		if err != nil {
			panic(err)
		}

		err = os.WriteFile("rows.go", []byte(mainFileContents), 0o600)
		if err != nil {
			panic(err)
		}

		testFileContents, err := g.GenerateGoTestFile()
		if err != nil {
			panic(err)
		}
		err = os.WriteFile("rows_test.go", []byte(testFileContents), 0o600)
		if err != nil {
			panic(err)
		}
	case "Tx":
		g := &gogenwrapper.T{
			PackageName:            packageName,
			StructNameWrapper:      "TxWrapper",
			StructNameInterceptor:  "TxInterceptor",
			FunctionNameWrap:       "WrapTx",
			FunctionTypeNamePrefix: "Tx_",
			InterfaceNameUnwrap:    "TxUnwrapper",
			FunctionNameUnwrap:     "UnwrapTx",
			BaseInterface:          new(driver.Tx),
		}

		mainFileContents, err := g.GenerateGoSourceFile()
		if err != nil {
			panic(err)
		}

		err = os.WriteFile("tx.go", []byte(mainFileContents), 0o600)
		if err != nil {
			panic(err)
		}

		testFileContents, err := g.GenerateGoTestFile()
		if err != nil {
			panic(err)
		}
		err = os.WriteFile("tx_test.go", []byte(testFileContents), 0o600)
		if err != nil {
			panic(err)
		}
	}
}
