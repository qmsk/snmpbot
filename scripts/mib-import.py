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
    }
    SYMBOL_CACHE = {}
    OID_CACHE = {
        ('SNMPv2-SMI', 'iso'): OID(1),
    }

    # list of explicitly supported syntax types
    SUPPORTED_SYNTAX = set([
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
    def lookupSyntax(self, name):
        mib = self.moduleName

        if name in self.objectTypes:
            log.debug("lookup type=%s: %r", name, self.objectTypes[name])

            return self.objectTypes[name]

        elif name in self.importTable:
            mib, symName = self.importTable[name]

            if (mib, symName) in self.SUPPORTED_SYNTAX:
                return '{mib}::{name}'.format(mib=mib, name=symName), None

            sym = self.lookupSymbol(mib, symName)

            log.debug("lookup type=%s => %s::%s: %r", name, mib, symName, sym)

            if sym['type'] != 'TypeDeclaration':
                raise TypeError("Invalid type=%s => %s::%s is a %s" % (name, mib, symName, sym['type']))

            syntax, syntax_spec = sym['syntax']
            syntax, syntax_mib = syntax

            # XXX: where do these get mapped?
            if syntax == 'Integer32' and syntax_spec:
                return 'ENUM', self.parseSyntaxEnum(syntax_spec)
            elif syntax == 'OctetString':
                return 'OCTET STRING', None
            elif syntax == 'ObjectIdentifier':
                return 'OBJECT IDENTIFIER', None
            else:
                # XXX: map remaining import types...
                return syntax, None
        else:
            raise ValueError("Unknown type=%s" % (name, ))

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

    def parseSyntax(self, value):
        if isinstance(value, str):
            return value, None
        else:
            syntax, options = value
            options = parseDeclArg(options)

            if syntax == 'INTEGER' and 'enumSpec' in options:
                return 'ENUM', self.parseSyntaxEnum(options['enumSpec'])
            elif syntax == 'INTEGER' and 'integerSubType' in options:
                return syntax, self.parseSyntaxOptions(options['integerSubType'][0])
            elif syntax == 'Integer32' and 'integerSubType' in options:
                return syntax, self.parseSyntaxOptions(options['integerSubType'][0])
            elif syntax == 'DisplayString' and 'octetStringSubType' in options:
                return syntax, self.parseSyntaxOptions(options['octetStringSubType'][0])
            elif syntax == 'OCTET STRING' and 'SnmpAdminString' in options: # XXX: should get resolved using lookupType!
                return syntax, self.parseSyntaxOptions(options['octetStringSubType'][0])
            else:
                return None, None

    def parseObjectSyntax(self, syntax):
        if 'SimpleSyntax' in syntax:
            return self.parseSyntax(syntax['SimpleSyntax'])
        elif 'ApplicationSyntax' in syntax:
            return self.parseSyntax(syntax['ApplicationSyntax'])
        elif 'row' in syntax:
            return self.lookupSyntax(syntax['row'])
        else:
            return None, None

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

        log.debug("register module=%s: oid=%s", name, self.moduleOID)

    def load_objectTypeClause_object(self, name, syntax, maxAccessPart, oid):
        oid = self.parseObjectIdentifier(oid['objectIdentifier'])

        syntax_name, syntax_options = self.parseObjectSyntax(syntax)

        # scalar objects
        if not syntax:
            log.warn("Object %s::%s has unsupported syntax: %r", self.moduleName, name, syntax)
        else:
            log.info("load object %s::%s@%s: %s, %s", self.moduleName, name, oid, syntax_name, syntax_options)

        object = {
            'Name': name,
            'OID': str(oid),
            'Syntax': syntax_name,
        }

        if syntax_options:
            object['SyntaxOptions'] = syntax_options

        if maxAccessPart and maxAccessPart['MaxAccessPart'] == 'not-accessible':
            object['NotAccessible'] = True

        self.objects.append(object)

    def load_objectTypeClause_table(self, name, syntax, oid):
        conceptualTable = parseDeclArg(syntax['conceptualTable'])
        oid = self.parseObjectIdentifier(oid['objectIdentifier'])

        entryType = conceptualTable['row']

        log.info("load table %s::%s@%s with entryType=%s", self.moduleName, name, oid, entryType)

        table = {
            'Name': name,
            'OID': str(oid),
        }

        self.tables.append(table)
        self.entryTable[entryType] = table

    def load_objectTypeClause_entry(self, name, syntax, index, oid):
        entryType = syntax['row']
        oid = self.parseObjectIdentifier(oid['objectIdentifier'])

        table = self.entryTable[entryType]
        entrySyntax = self.entryTypes[entryType]

        if index and 'INDEX' in index:
            indexSyntax = index['INDEX']
        else:
            # TODO: AUGEMENTS?
            log.warn("Entry %s::%s missing INDEX: %r", self.moduleName, name, index)
            return

        log.info("load entry %s::%s@%s for table=%s: index=%s syntax=%s", self.moduleName, name, oid, table['Name'], indexSyntax, entrySyntax)

        table['IndexObjects'] = [self.formatObject(name) for i, name in indexSyntax]
        table['EntryObjects'] = [self.formatObject(name) for name, syntax in entrySyntax] # TODO: only if object is accessible?

    def load_objectTypeClause(self, name, syntax, units, maxAccessPart, status, description, reference, augmention, index, defval, oid):
        # table, entry, object with custom syntax, or object with built-in syntax?
        if 'conceptualTable' in syntax:
            return self.load_objectTypeClause_table(name, syntax, oid)
        elif 'row' in syntax and syntax['row'] in self.entryTypes:
            return self.load_objectTypeClause_entry(name, syntax, index, oid)
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
            args = parseDeclArgs(args)

            if type == 'typeDeclaration':
                log.debug("load mib=%s <%s>%s: %s", moduleName, type, name, args)

                ctx.load_typeDeclaration(name, *args)

        # objects pass
        for type, name, *args in declarations:
            args = parseDeclArgs(args)

            print("\t{type:<20} {mib:>18}::{name:<30} = {args!r}".format(type=type, mib=mib, name=name, args=args))

            # generate
            if type == 'moduleIdentityClause':
                log.debug("load mib=%s <%s>%s: %s", moduleName, type, name, args)

                ctx.load_moduleIdentityClause(name, *args)

            elif type == 'objectTypeClause':
                log.debug("load mib=%s <%s>%s: %s", moduleName, type, name, args)

                ctx.load_objectTypeClause(name, *args)

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

    print("Compiled MIBs to {path}:".format(path=args.output_path))
    for mib, status in compile_status.items():
        print("\t{mib}.json: {status}".format(mib=mib, status=status))

    return 0

if __name__ == '__main__':
    import sys

    sys.exit(main())
