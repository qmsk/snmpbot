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

log = logging.getLogger('mib-import')

def parseDeclArg(arg):
    if arg is None:
        return None
    elif isinstance(arg, tuple):
        if len(arg) == 2:
            attr, value = arg

            return {attr: value}
        else:
            attr, *values = arg

            return {attr: parseDeclArg(values)}
    else:
        return arg

def parseDeclArgs(args):
    return tuple(parseDeclArg(arg) for arg in args)

def parseOIDPart(part):
    if isinstance(part, tuple):
        name, id = part
    else:
        id = part

    return id

class OID:
    def __init__(self, *ids):
        self.ids = ids

    def extend(self, *ids):
        return OID(*(self.ids + ids))

    def __str__(self):
        return '.' + '.'.join(str(x) for x in self.ids)

class Context:
    IMPORT_TABLE = {
        'iso': ('SNMPv2-SMI', 'iso'),
        'Counter': ('SNMPv2-SMI', 'Counter32'),
        'Gauge': ('SNMPv2-SMI', 'Gauge32'),
    }
    SYMBOL_CACHE = {}
    OID_CACHE = {
        ('SNMPv2-SMI', 'iso'): OID(1),
    }

    # list of explicitly supported syntax types
    SIMPLE_SYNTAX = set([
        'INTEGER',
        'OCTET STRING',
        'OBJECT IDENTIFIER',
    ])
    APPLICATION_SYNTAX = set([
        'IpAddress',
        'TimeTicks',
        'Opaque',

        'Counter32',
        'Gauge32',
        'Integer32',
        'Unsigned32',

        'Counter64',
    ])
    SUPPORTED_SYNTAX = set([
        ('SNMP-FRAMEWORK-MIB', 'SnmpAdminString'),
        ('SNMPv2-TC', 'DisplayString'),
        ('SNMPv2-TC', 'MacAddress'),
        ('SNMPv2-TC', 'PhysAddress'),
        ('Q-BRIDGE-MIB', 'PortList'),
        ('BRIDGE-MIB', 'BridgeId'),
    ])

    @classmethod
    def loadImports(cls, imports, convertTable):
        importTable = dict(cls.IMPORT_TABLE)

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

        self.objectTypes = {} # SYNTAX row=* => TEXTUAL-CONVENTION<...>
        self.entryTypes = {} # SYNTAX row= => <SEQUENCE>(name, syntax)]
        self.entryTable = {} # mapping from Entry type => Table name

        self.moduleOID = None
        self.objects = []
        self.tables = []

    def lookupSymbol(self, mib, name):
        sym = self.symbolCache.get((mib, name))

        if not sym:
            if mib == self.moduleName and name in self.importTable:
                mib, name = self.importTable[name]

            sym = self.symbolCache[mib, name] = self.symbolTable[mib][name.replace('-', '_')]

        return sym

    def resolveName(self, name):
        if name in self.importTable:
            mib, name = self.importTable[name]
        else:
            mib = self.moduleName

        return mib, name

    def resolveOID(self, mib, name, *ids):
        if mib == self.moduleName:
            mib, name = self.resolveName(name)

        oid = self.oidCache.get((mib, name))

        if not oid:
            sym = self.lookupSymbol(mib, name)

            parent, *parent_ids = sym['oid']
            parent_name, parent_mib = parent

            oid = self.oidCache[mib, name] = self.resolveOID(parent_mib, parent_name, *parent_ids)

        oid = oid.extend(*ids)

        return oid

    # @return [syntax, options]
    def lookupSyntax(self, syntaxName, options=None):
        if syntaxName in self.objectTypes:
            syntax, options = self.objectTypes[syntaxName]

            log.debug("lookup type=%s: %s %r", syntaxName, syntax, options)

            return syntax, options

        mib = self.moduleName
        name = syntaxName

        if name in self.importTable:
            mib, name = self.importTable[name]

        if (mib, name) in self.SUPPORTED_SYNTAX:
            return '{mib}::{name}'.format(mib=mib, name=name), None

        sym = self.lookupSymbol(mib, name)

        if not sym:
            raise ValueError("Unknown syntax: {mib}::{name}".format(mib=mib, name=name))

        if sym['type'] == 'TypeDeclaration':
            log.debug("lookup type=%s => %s::%s: %r", syntaxName, mib, name, sym)
        else:
            raise ValueError("Invalid type=%s => %s::%s is a %s" % (syntaxName, mib, name, sym['type']))

        syntax, syntax_spec = sym['syntax']
        syntax, syntax_mib = syntax

        # these get mapped stupidly by the symtable codegen
        if syntax == 'Integer32' and syntax_spec:
            return 'ENUM', self.parseSyntaxEnum(syntax_spec)
        elif syntax == 'OctetString':
            return 'OCTET STRING', options
        elif syntax == 'ObjectIdentifier':
            return 'OBJECT IDENTIFIER', options
        elif syntax in self.SIMPLE_SYNTAX:
            return syntax, options
        elif syntax in self.APPLICATION_SYNTAX:
            return syntax, options
        elif (syntax_mib, syntax) in self.SUPPORTED_SYNTAX:
            return '{mib}::{name}'.format(mib=syntax_mib, name=syntax), options
        else:
            raise ValueError("Invalid syntax={syntaxName} => {mib}::{name} => {syntax}".format(syntaxName=syntaxName, mib=mib, name=name, syntax=syntax))

    def parseObjectIdentifier(self, objectIdentifier):
        parent, *ids = objectIdentifier

        ids = [parseOIDPart(id) for id in ids]

        return self.resolveOID(self.moduleName, parent, *ids)

    def parseSyntaxOptions(self, options):
        if len(options) == 1:
            min = max = options[0]
        elif len(options) == 2:
            min, max = options
        else:
            raise ValueError("Invalid SYNTAX options: {options}".format(options=options))

        return {'Min': min, 'Max': max}

    def parseSyntaxEnum(self, spec):
        return [{'Value': value, 'Name': name} for name, value in spec]

    def parseSyntaxBits(self, spec):
        return [{'Bit': bit, 'Name': name} for name, bit in spec]

    def parseSyntax(self, value):
        if isinstance(value, str):
            syntax = value
            options = None
        else:
            syntax, options = value
            options = parseDeclArg(options)

        if not options:
            syntax_options = None
        elif 'enumSpec' in options:
            return 'ENUM', self.parseSyntaxEnum(options['enumSpec'])
        elif 'integerSubType' in options:
            syntax_options = self.parseSyntaxOptions(options['integerSubType'][0])
        elif 'octetStringSubType' in options:
            syntax_options = self.parseSyntaxOptions(options['octetStringSubType'][0])
        else:
            syntax_options = None

        if syntax in self.SIMPLE_SYNTAX:
            return syntax, syntax_options
        elif syntax in self.APPLICATION_SYNTAX:
            return syntax, syntax_options
        else:
            return self.lookupSyntax(syntax, syntax_options)

    def parseObjectSyntax(self, syntax):
        if 'SimpleSyntax' in syntax:
            return self.parseSyntax(syntax['SimpleSyntax'])
        elif 'ApplicationSyntax' in syntax:
            return self.parseSyntax(syntax['ApplicationSyntax'])
        elif 'row' in syntax:
            return self.lookupSyntax(syntax['row'])
        elif 'BITS' in syntax:
            return 'BITS', self.parseSyntaxBits(syntax['BITS'])
        else:
            raise ValueError("Invalid syntax for object: {syntax}".format(syntax=syntax))

    def formatObject(self, name):
        mib, name = self.resolveName(name)

        return '{mib}::{name}'.format(mib=mib, name=name)

    def load_typeDeclaration_textualConvention(self, name, display, status, description, reference, syntax):
        syntax_name, options = self.objectTypes[name] = self.parseObjectSyntax(syntax)

        log.debug("register object type=%s: %s, %r", name, syntax_name, options)

    def load_typeDeclaration(self, name, typeDeclarationRHS):
        typeDeclarationRHS = parseDeclArg(typeDeclarationRHS['typeDeclarationRHS'])

        if isinstance(typeDeclarationRHS, list):
            self.load_typeDeclaration_textualConvention(name, *parseDeclArgs(typeDeclarationRHS))

        elif 'SEQUENCE' in typeDeclarationRHS:
            entrySyntax = self.entryTypes[name] = typeDeclarationRHS['SEQUENCE']

            log.debug("register entry type=%s: %s", name, entrySyntax)

        else: # XXX: This is an odd case for thigns like TOKEN-RING-RMON-MIB::MacAddress
            syntax_name, options = self.objectTypes[name] = self.parseObjectSyntax(typeDeclarationRHS)

            log.debug("register object type=%s: %s, %s", name, syntax_name, options)

    def load_moduleIdentityClause(self, name, lastUpdated, organization, contactInfo, description, revisions, oid):
        self.moduleOID = self.parseObjectIdentifier(oid['objectIdentifier'])

        log.info("load mib %s@%s", name, self.moduleOID)

    def load_objectTypeClause_object(self, name, syntax, maxAccessPart, oid):
        oid = self.parseObjectIdentifier(oid['objectIdentifier'])

        syntax_name, syntax_options = self.parseObjectSyntax(syntax)

        # scalar objects
        if not syntax:
            log.warn("Object %s::%s has unsupported syntax: %r", self.moduleName, name, syntax)

        object = {
            'Name': name,
            'OID': str(oid),
            'Syntax': syntax_name,
        }

        if syntax_options:
            object['SyntaxOptions'] = syntax_options

        if maxAccessPart and maxAccessPart['MaxAccessPart'] == 'not-accessible':
            object['NotAccessible'] = True

        log.info("load object %s::%s@%s: %r", self.moduleName, name, oid, object)

        self.objects.append(object)

    def load_objectTypeClause_table(self, name, syntax, oid):
        conceptualTable = parseDeclArg(syntax['conceptualTable'])
        oid = self.parseObjectIdentifier(oid['objectIdentifier'])

        entryType = conceptualTable['row']

        table = {
            'Name': name,
            'OID': str(oid),
        }

        log.info("load table %s::%s@%s with entryType=%s", self.moduleName, name, oid, entryType)

        self.tables.append(table)
        self.entryTable[entryType] = table

    def load_objectTypeClause_entry(self, name, syntax, augmention, index, oid):
        entryType = syntax['row']
        oid = self.parseObjectIdentifier(oid['objectIdentifier'])

        table = self.entryTable[entryType]
        entrySyntax = self.entryTypes[entryType]

        table['EntryName'] = name

        if augmention:
            entryName = self.formatObject(augmention)

            table['AugmentsEntry'] = entryName

        elif index and 'INDEX' in index:
            indexSyntax = index['INDEX']

            table['IndexObjects'] = [self.formatObject(name) for i, name in indexSyntax]
        else:
            raise ValueError("Missing AUGMENTS/INDEX: %r %r", augmention, index)

        table['EntryObjects'] = [self.formatObject(name) for name, syntax in entrySyntax] # TODO: only if object is accessible?

        log.info("load entry %s::%s@%s for table=%s: %r", self.moduleName, name, oid, table['Name'], table)

    # load objects once all types are registered
    def load_objectTypeClause(self, name, syntax, units, maxAccessPart, status, description, reference, augmention, index, defval, oid):
        if 'conceptualTable' in syntax:
            return self.load_objectTypeClause_table(name, syntax, oid)
        elif 'row' in syntax and syntax['row'] in self.entryTypes:
            return self.load_objectTypeClause_entry(name, syntax, augmention, index, oid)
        else:
            return self.load_objectTypeClause_object(name, syntax, maxAccessPart, oid)

class CodeGen(pysmi.codegen.base.AbstractCodeGen):
    def genCode(self, ast, symbolTable, **kwargs):
        moduleName, moduleOID, imports, declarations = ast
        moduleIdentity = None

        ctx = Context(moduleName,
            symbolTable     = symbolTable,
            imports         = imports,
            convertTable    = self.convertImportv2,
        )

        print("{mib}:".format(mib=moduleName))

        # imports
        for mib, names in imports.items():
            for name in names:
                print("\t{type:<20} {name:<30} = {mib}::{name}".format(type='import', mib=mib, name=name))

        # symbols
        for mib, symbols in symbolTable.items():
            for name, sym in symbols.items():
                print("\t{type:<20} {mib:>18}::{name:<30} = {attrs!r}".format(type='symbol', mib=mib, name=name, attrs=sym))

        # types pass
        for type, name, *args in declarations:
            try:
                args = parseDeclArgs(args)

                if type == 'moduleIdentityClause':
                    log.debug("load mib=%s <%s>%s: %s", moduleName, type, name, args)

                    ctx.load_moduleIdentityClause(name, *args)

                elif type == 'typeDeclaration':
                    log.debug("load mib=%s <%s>%s: %s", moduleName, type, name, args)

                    ctx.load_typeDeclaration(name, *args)
            except Exception as exc:
                log.exception("Failed to load {type} {mib}::{name}: {exc}".format(type=type, mib=moduleName, name=name, exc=exc))
                raise exc

        # objects pass
        for type, name, *args in declarations:
            args = parseDeclArgs(args)

            print("\t{type:<20} {mib:>18}::{name:<30} = {args!r}".format(type=type, mib=moduleName, name=name, args=args))

            try:
                if type == 'objectTypeClause':
                    log.debug("load mib=%s <%s>%s: %s", moduleName, type, name, args)

                    ctx.load_objectTypeClause(name, *args)
            except Exception as exc:
                log.exception("Failed to load {type} {mib}::{name}: {exc}".format(type=type, mib=moduleName, name=name, exc=exc))
                raise exc

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
        mibdata = json.dumps(out, indent=2, sort_keys=True)

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

    print("Compiled MIBs to {path}:".format(path=args.output_path))
    for mib, status in compile_status.items():
        print("\t{mib}.json: {status}".format(mib=mib, status=status))

    return 0

if __name__ == '__main__':
    import sys

    sys.exit(main())
