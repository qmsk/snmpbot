#!/usr/bin/env python3

import argparse
import pysmi.compiler
import pysmi.codegen
import pysmi.debug
import pysmi.parser
import pysmi.reader
import pysmi.searcher
import pysmi.writer

"""
Generate github.com/qmsk/snmpbot JSON MIB files from ASN.1 SMI sources.
"""

__version__ = '0.1'

MIB_PATH = [
    '/usr/share/mibs/',
    '/usr/share/snmp/mibs',
    '/usr/local/share/mibs',
]

def build_compiler(args):
    compiler = pysmi.compiler.MibCompiler(
        parser  = pysmi.parser.SmiStarParser(),
        codegen = pysmi.codegen.JsonCodeGen(),
        writer  = pysmi.writer.FileWriter(args.output_path).setOptions(suffix='.json'),
    )

    for path in args.mib_path:
        compiler.addSources(pysmi.reader.FileReader(path, recursive=True))
    for url in args.mib_url:
        compiler.addSources(*pysmi.reader.getReadersFromUrls(url))

    compiler.addSearchers(pysmi.searcher.StubSearcher(*pysmi.codegen.JsonCodeGen.baseMibs))

    return compiler

def main():
    parser = argparse.ArgumentParser(
        description = __doc__,
    )
    parser.add_argument('--debug', nargs='?', const='all')
    parser.add_argument('--mib-path', nargs='*', metavar='PATH', default=MIB_PATH)
    parser.add_argument('--mib-url', nargs='*', metavar='URL', default=[])
    parser.add_argument('--output-path', metavar='PATH', required=True)
    parser.add_argument('--rebuild', action='store_true')
    parser.add_argument('mibs', metavar='MIB', nargs='+')

    args = parser.parse_args()

    if args.debug:
        pysmi.debug.setLogger(pysmi.debug.Debug(args.debug))

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
