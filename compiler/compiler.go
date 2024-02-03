package compiler

import (
	"blue/ast"
	"fmt"
	"log"
)

func compileProgram(program *ast.Program) string {
	// TODO: Need to add imports, file organization, anything else

	var result string
	for _, stmt := range program.Statements {
		result = Compile(stmt)
		// TODO: Figure out Return vs error?
		// switch result := result.(type) {
		// case *object.ReturnValue:
		// 	return result.Value
		// case *object.Error:
		// 	return result
		// }
	}
	return result
}

// TODO: Make ast.Program work the way we do in eval
func Compile(node ast.Node) string {
	log.Printf("node `%s` (%T)", node.String(), node)
	switch node := node.(type) {
	case *ast.Program:
		return compileProgram(node)
	case *ast.BreakStatement:
	case *ast.ContinueStatement:
	case *ast.ExpressionStatement:
		obj := Compile(node.Expression)
		return obj
	case *ast.Identifier:
	case *ast.IntegerLiteral:
		return fmt.Sprintf("&object.Integer{Value: %d}", node.Value)
	case *ast.BigIntegerLiteral:
	case *ast.HexLiteral:
		return fmt.Sprintf("&object.UInteger{Value: %x}", node.Value)
	case *ast.OctalLiteral:
	case *ast.BinaryLiteral:
	case *ast.UIntegerLiteral:
	case *ast.FloatLiteral:
		return fmt.Sprintf("&object.Float{Value: %f}", node.Value)
	case *ast.BigFloatLiteral:
	case *ast.Boolean:
	case *ast.PrefixExpression:
	case *ast.InfixExpression:
	case *ast.PostfixExpression:
	case *ast.BlockStatement:
	case *ast.IfExpression:
	case *ast.ReturnStatement:
	case *ast.ValStatement:
	case *ast.VarStatement:
	case *ast.FunctionLiteral:
	case *ast.FunctionStatement:
	case *ast.CallExpression:
	case *ast.RegexLiteral:
	case *ast.StringLiteral:
	case *ast.ExecStringLiteral:
	case *ast.ListLiteral:
	case *ast.IndexExpression:
	case *ast.MapLiteral:
	case *ast.AssignmentExpression:
	case *ast.ForExpression:
	case *ast.ListCompLiteral:
	case *ast.MapCompLiteral:
	case *ast.SetCompLiteral:
	case *ast.MatchExpression:
	case *ast.Null:
	case *ast.SetLiteral:
	case *ast.ImportStatement:
	case *ast.TryCatchStatement:
	case *ast.EvalExpression:
	case *ast.SpawnExpression:
	case *ast.DeferExpression:
	case *ast.SelfExpression:
	default:
		if node == nil {
			// Just want to get rid of this in my output
			return "TODO: ?"
		}
		fmt.Printf("Handle this type: %T\n", node)
	}
	return "nil"
}
