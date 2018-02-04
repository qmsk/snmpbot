#!/usr/bin/env python3

import argparse
import json
import pprint
import pysmi.borrower
import pysmi.compiler
import pysmi.codegen
import pysmi.codegen.base
import pysmi.debug
import pysmi.mibinfo
import pysmi.parser
import pysmi.reader
import pysmi.searcher
import pysmi.writer
import logging

"""
Generate github.com/qmsk/snmpbot JSON MIB files from ASN.1 SMI sources.
"""

__version__ = '0.1'

MIB_PATH = [
    '/usr/share/mibs/',
    '/usr/share/snmp/mibs',
    '/usr/local/share/mibs',
]
MIB_URLS = [
    'http://mibs.snmplabs.com/asn1/@mib@',
]
MIB_BORROWERS = [
    'http://mibs.snmplabs.com/pysnmp/notexts/@mib@',
]

log = logging.getLogger('main')

def parseDeclAttrs(args):
    attrs = {}

    for arg in args:
        if arg is None:
            continue
        elif isinstance(arg, tuple):
            attr, *values = arg

            if len(values) == 1 and not isinstance(values[0], tuple):
                attrs[attr] = values[0]
            else:
                attrs[attr] = parseDeclAttrs(values)
        elif isinstance(arg, list):
            if len(arg) == 0:
                # ???
                continue
            else:
                raise ValueError("Unexpected list: %r", arg)
        else:
            attrs[arg] = None

    return attrs

class OID:
    def __init__(self, *ids):
        self.ids = ids

    def extend(self, *ids):
        return OID(*(self.ids + ids))

    def __str__(self):
        return '.' + '.'.join(str(x) for x in self.ids)

class Context:
    SYMBOL_CACHE = {}
    OID_CACHE = {
        ('SNMPv2-SMI', 'iso'): OID(1),
    }

    # list of explicitly supported syntax types
    SUPPORTED_SYNTAX = set([
        ('Q-BRIDGE-MIB', 'PortList'),
        ('BRIDGE-MIB', 'BridgeId'),
    ])

    @classmethod
    def loadImports(cls, imports, convertTable):
        importTable = {}

        for mib, names in imports.items():
            convertMap = convertTable.get(mib)

            for name in names:
                if convertMap and name in convertMap:
                    converted = importTable[name] = next(convertMap[name]) # XXX: why list?

                    log.debug("convert import %s::%s => %s::%s" , mib, name, *converted)
                else:
                    importTable[name] = (mib, name)

        return importTable

    def __init__(self, moduleName, symbolTable, imports, convertTable):
        self.moduleName = moduleName
        self.symbolTable = symbolTable
        self.importTable = self.loadImports(imports, convertTable=convertTable)
        self.symbolCache = dict(self.SYMBOL_CACHE)
        self.oidCache = dict(self.OID_CACHE)

        self.types = {}
        self.objects = []
        self.tables = []

    def lookupSymbol(self, mib, name):
        sym = self.symbolCache.get((mib, name))

        if not sym:
            if mib == self.moduleName and name in self.importTable:
                mib, name = self.importTable[name]

            sym = self.symbolCache[mib, name] = self.symbolTable[mib][name.replace('-', '_')]

        return sym

    def lookup(self, mib, name, id=None):
        oid = self.oidCache.get((mib, name))

        if not oid:
            sym = self.lookupSymbol(mib, name)

            parent, parent_id = sym['oid']
            parent_name, parent_mib = parent

            oid = self.oidCache[mib, name] = self.lookup(parent_mib, parent_name, parent_id)

        if id:
            oid = oid.extend(id)

        return oid

    # returns raw syntax
    def lookupType(self, name):
        mib = self.moduleName

        if name in self.types:
            return self.types[name]
        elif name in self.importTable:
            # XXX: the lookup should happen via the symbolTable with imports?
            mib, symName = self.importTable[name]
            sym = self.lookupSymbol(mib, symName)

            if sym['type'] != 'TypeDeclaration':
                raise TypeError("Invalid type=%s => %s::%s is a %s" % (name, mib, symName, sym['type']))

            return sym['syntax']
        else:
            raise ValueError("Unknown type=%s" % (name, ))

    def parseSyntaxOptions(self, name, options):
        if len(options) == 1:
            min = max = options[0]
        elif len(options) == 2:
            min, max = options
        else:
            raise ValueError("Invalid SYNTAX options for %s::%s: %s" % (self.moduleName, name, options))

        return {'Min': min, 'Max': max}

    def parseSyntax(self, name, value):
        syntax = None
        options = None

        if isinstance(value, str):
            syntax = value
        elif 'INTEGER' in value and 'enumSpec' in value:
            syntax = 'ENUM'
            options = [{'Value': value, 'Name': name} for name, value in value['enumSpec']]
        elif 'INTEGER' in value and 'integerSubType' in value:
            syntax = 'INTEGER'
            options = self.parseSyntaxOptions(name, value['integerSubType'][0])
        elif 'Integer32' in value and 'integerSubType' in value:
            syntax = 'Integer32'
            options = self.parseSyntaxOptions(name, value['integerSubType'][0])
        elif 'DisplayString' in value:
            syntax = 'DisplayString'
            options = self.parseSyntaxOptions(name, value['octetStringSubType'][0])
        elif 'OCTET STRING' in value:
            syntax = 'OCTET STRING'
            options = self.parseSyntaxOptions(name, value['octetStringSubType'][0])
        elif len(value) == 1:
            # return simple syntax key
            for key in value:
                return key, None

        return syntax, options

class CodeGen(pysmi.codegen.base.AbstractCodeGen):
    # register a TEXTUAL-CONVENTION as an alias for some SYNTAX
    def registerType(self, ctx, name, attrs):
        typeAttrs = attrs.get('typeDeclarationRHS')

        log.debug("register type=%s: %s", name, typeAttrs)

        ctx.types[name] = typeAttrs

    def genObject(self, ctx, oid, name, attrs):
        rawSyntax = None

        if 'row' in attrs:
            # row => any UpperCase SYNTAX
            typeAttrs = ctx.lookupType(attrs['row'])
        else:
            typeAttrs = attrs

        if 'SimpleSyntax' in typeAttrs:
            rawSyntax = typeAttrs['SimpleSyntax']
        elif 'ApplicationSyntax' in typeAttrs:
            rawSyntax = typeAttrs['ApplicationSyntax']
        else:
            log.warn("Skip %s::%s without scalar syntax", ctx.moduleName, name)
            return

        syntax, syntax_options = ctx.parseSyntax(name, rawSyntax)

        if not syntax:
            log.warn("Skip %s::%s with unsupported syntax: %s", ctx.moduleName, name, rawSyntax)
            return

        object = {
            'Name': name,
            'OID': str(oid),
            'Syntax': syntax,
        }

        if syntax_options:
            object['SyntaxOptions'] = syntax_options

        if attrs.get('MaxAccessPart') == 'not-accessible':
            object['NotAccessible'] = True

        ctx.objects.append(object)

    def genTable(self, ctx, oid, name, attrs):
        table = {
            'Name': name,
            'OID': str(oid),
        }

        ctx.tables.append(table)

    def genCode(self, ast, symbolTable, **kwargs):
        moduleName, moduleOID, imports, declarations = ast
        moduleIdentity = None

        print(json.dumps(dict(
            moduleName  = moduleName,
            moduleOID   = moduleOID,
        #    imports     = imports,
        #    declarations=declarations,
        #    symbolTable = symbolTable,
        #    kwargs=kwargs,
        ), indent=2))

        ctx = Context(moduleName,
            symbolTable     = symbolTable,
            imports         = imports,
            convertTable    = self.convertImportv2,
        )

        print("{mib}:".format(mib=moduleName))

        # type pass
        for type, name, *args in declarations:
            if type == 'typeDeclaration':
                self.registerType(ctx, name, parseDeclAttrs(args))

        # objects pass
        for type, name, *args in declarations:
            # parse
            log.debug("parse mib=%s decl <%s>%s: %s", moduleName, type, name, args)

            attrs = parseDeclAttrs(args)

            # dump
            if 'objectIdentifier' in attrs:
                ref = attrs['objectIdentifier']

                if len(ref) == 1:
                    parent_name, = ref
                    id = None
                elif len(ref) == 2:
                    parent_name, id = ref
                else:
                    raise ValueError("Invalid objectIdentifier for %s::%s: %s", ctx.modueName, name, ref)

                oid = ctx.lookup(moduleName, parent_name, id)
            else:
                oid = None

            print("\t{type:15} {name:30} {oid}".format(type=type, name=name, oid=oid))

            for attr, value in attrs.items():
                print("\t\t{attr:20}: {value}".format(attr=attr, value=value))

            # generate
            if type == 'moduleIdentityClause':
                moduleOID = oid

            elif type == 'objectTypeClause':
                self.genObject(ctx, oid, name, attrs)

        out = {
            'Name': moduleName,
            'Objects': ctx.objects,
            'Tables': ctx.tables,
        }

        if moduleOID:
            out['OID'] = str(moduleOID)

        mibinfo = pysmi.mibinfo.MibInfo(
            oid         = moduleOID,
            identity    = moduleIdentity,
            name        = moduleName,
        )
        mibdata = json.dumps(out, indent=2)

        return mibinfo, mibdata

def build_compiler(args):
    codegen = CodeGen()

    compiler = pysmi.compiler.MibCompiler(
        parser  = pysmi.parser.SmiStarParser(),
        codegen = codegen,
        writer  = pysmi.writer.FileWriter(args.output_path).setOptions(suffix='.json'),
    )

    for path in args.mib_path:
        compiler.addSources(pysmi.reader.FileReader(path, recursive=True))
    for url in args.mib_url:
        compiler.addSources(*pysmi.reader.getReadersFromUrls(url))
    for url in args.mib_borrowers:
        for reader in pysmi.reader.getReadersFromUrls(url):
            compiler.addBorrowers(pysmi.borrower.PyFileBorrower(reader))

    compiler.addSearchers(pysmi.searcher.StubSearcher(*codegen.baseMibs))

    return compiler

def main():
    parser = argparse.ArgumentParser(
        description = __doc__,
    )
    parser.add_argument('--debug', action='store_true')
    parser.add_argument('--verbose', action='store_true')
    parser.add_argument('--pysmi-debug', nargs='?', const='all')
    parser.add_argument('--mib-path', nargs='*', metavar='PATH', default=MIB_PATH)
    parser.add_argument('--mib-url', nargs='*', metavar='URL', default=MIB_URLS)
    parser.add_argument('--mib-borrowers', nargs='*', metavar='URL', default=MIB_BORROWERS)
    parser.add_argument('--output-path', metavar='PATH', required=True)
    parser.add_argument('--rebuild', action='store_true')
    parser.add_argument('mibs', metavar='MIB', nargs='+')

    args = parser.parse_args()

    logging.basicConfig()

    if args.debug:
        log.setLevel(logging.DEBUG)
    elif args.verbose:
        log.setLevel(logging.INFO)
    else:
        log.setLevel(logging.WARN)

    if args.pysmi_debug:
        pysmi.debug.setLogger(pysmi.debug.Debug(args.pysmi_debug))

    compiler = build_compiler(args)
    compile_status = compiler.compile(*args.mibs,
        rebuild = args.rebuild,
    )

    for mib, status in compile_status.items():
        print(mib, status)

    return 0

if __name__ == '__main__':
    import sys

    sys.exit(main())
