import XCTest
import SwiftTreeSitter
import TreeSitterLuax

final class TreeSitterLuaxTests: XCTestCase {
    func testCanLoadGrammar() throws {
        let parser = Parser()
        let language = Language(language: tree_sitter_luax())
        XCTAssertNoThrow(try parser.setLanguage(language),
                         "Error loading LuaX grammar")
    }
}
