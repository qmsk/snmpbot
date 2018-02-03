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

def parseDeclArgs(args):
    attrs = {}

    for arg in args:
        if arg is None:
            continue
        elif isinstance(arg, tuple):
            attr, *values = arg

            if len(values) == 1:
                attrs[attr] = values[0]
            else:
                attrs[attr] = parseDeclArgs(values)
        elif isinstance(arg, list):
            if len(arg) == 0:
                # ???
                continue
            else:
                raise ValueError("Unexpected list: %r", arg)
        else:
            attrs[arg] = None

    return attrs

def parseDecl(decl):
    type, name, *args = decl

    return type, name, parseDeclArgs(args)

class OID:
    def __init__(self, *ids):
        self.ids = ids

    def extend(self, *ids):
        return OID(*(self.ids + ids))

    def __str__(self):
        return '.' + '.'.join(str(x) for x in self.ids)

class Context:
    SYMBOL_CACHE = {
        ('SNMPv2-SMI', 'iso'): OID(1),
    }

    SIMPLE_SYNTAX = set((
        'TimeTicks',
        'Counter32',
        'OBJECT IDENTIFIER',
    ))

    def __init__(self, moduleName, symbolTable, imports):
        self.moduleName = moduleName
        self.symbolTable = symbolTable
        self.importTable = {name: mib for mib, names in imports.items() for name in names}
        self.symbolCache = dict(self.SYMBOL_CACHE)

        self.objects = []
        self.tables = []

    def lookup(self, mib, name, id=None):
        oid = self.symbolCache.get((mib, name))

        if not oid:
            if mib == self.moduleName and name in self.importTable:
                mib = self.importTable[name]

            sym = self.symbolTable[mib][name.replace('-', '_')]
            parent, parent_id = sym['oid']
            parent_name, parent_mib = parent

            oid = self.symbolCache[mib, name] = self.lookup(parent_mib, parent_name, parent_id)

        if id:
            oid = oid.extend(id)

        return oid

    def parseSyntax(self, name, value):
        syntax = None
        options = None

        if isinstance(value, str):
            syntax = value
        elif 'INTEGER' in value and 'enumSpec' in value:
            syntax = 'ENUM'
            options = [{'Value': value, 'Name': name} for name, value in value['enumSpec']]
        elif 'DisplayString' in value:
            syntax = 'DisplayString'
            options = value['octetStringSubType']
            size_min, size_max = options[0]
            options = {'SizeMin': size_min, 'SizeMax': size_max}
        elif len(value) == 1:
            # return simple syntax key
            for key in value:
                return key, None

        return syntax, options

class CodeGen(pysmi.codegen.base.AbstractCodeGen):
    def genObject(self, ctx, oid, name, attrs):
        rawSyntax = attrs.get('SimpleSyntax') or attrs.get('ApplicationSyntax')

        if not rawSyntax:
            log.warn("Skip %s::%s without syntax", ctx.moduleName, name)
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
            symbolTable = symbolTable,
            imports     = imports,
        )

        print("{mib}:".format(mib=moduleName))

        for decl in declarations:
            # parse
            log.debug("parse mib=%s decl: %s", moduleName, decl)

            type, name, attrs = parseDecl(decl)

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

            elif type == 'objectTypeClause' and 'conceptualTable' in attrs:
                self.genTable(ctx, oid, name, attrs)

            elif type == 'objectTypeClause' and 'row' in attrs:
                pass

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
