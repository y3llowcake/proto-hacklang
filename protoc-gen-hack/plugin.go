package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"github.com/golang/protobuf/proto"
	desc "github.com/golang/protobuf/protoc-gen-go/descriptor"
	ppb "github.com/y3llowcake/proto-hack/third_party/gen-src/github.com/golang/protobuf/protoc-gen-go/plugin"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	genDebug      = false
	libNs         = "\\Protobuf"
	libNsInternal = libNs + "\\Internal"
)

func main() {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(os.Stdin)
	if err != nil {
		panic(fmt.Errorf("error reading from stdin: %v", err))
	}
	out, err := codeGenerator(buf.Bytes())
	if err != nil {
		panic(err)
	}
	os.Stdout.Write(out)
}

func codeGenerator(b []byte) ([]byte, error) {
	req := ppb.CodeGeneratorRequest{}
	err := proto.Unmarshal(b, &req)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling CodeGeneratorRequest: %v", err)
	}
	resp := gen(&req)
	out, err := proto.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("error marshaling CodeGeneratorResponse: %v", err)
	}
	return out, nil
}

func gen(req *ppb.CodeGeneratorRequest) *ppb.CodeGeneratorResponse {
	resp := &ppb.CodeGeneratorResponse{}
	fileToGenerate := map[string]bool{}
	for _, f := range req.FileToGenerate {
		fileToGenerate[f] = true
	}
	genService := strings.Contains(req.GetParameter(), "plugins=grpc")
	genService = genService || strings.Contains(req.GetParameter(), "plugin=grpc")

	rootns := NewEmptyNamespace()
	for _, fdp := range req.ProtoFile {
		if *fdp.Syntax != "proto3" {
			panic(fmt.Errorf("unsupported syntax: %s in file %s", *fdp.Syntax, *fdp.Name))
		}
		rootns.Parse(fdp)
		// panic(rootns.PrettyPrint())

		if !fileToGenerate[*fdp.Name] {
			continue
		}
		f := &ppb.CodeGeneratorResponse_File{}

		fext := filepath.Ext(*fdp.Name)
		fname := strings.TrimSuffix(*fdp.Name, fext) + "_proto.php"
		f.Name = proto.String(fname)

		b := &bytes.Buffer{}
		w := &writer{b, 0}
		writeFile(w, fdp, rootns, genService)
		f.Content = proto.String(b.String())
		resp.File = append(resp.File, f)
	}
	return resp
}

func writeFile(w *writer, fdp *desc.FileDescriptorProto, rootNs *Namespace, genService bool) {
	packageParts := strings.Split(fdp.GetPackage(), ".")
	ns := rootNs.FindFullyQualifiedNamespace("." + fdp.GetPackage())
	if ns == nil {
		panic("unable to find namespace for: " + fdp.GetPackage())
	}

	// File header.
	w.p("<?hh // strict")
	w.p("namespace %s;", strings.Join(packageParts, "\\"))
	w.ln()
	w.p("// Generated by the protocol buffer compiler.  DO NOT EDIT!")
	w.p("// Source: %s", fdp.GetName())
	w.ln()

	// Top level enums.
	for _, edp := range fdp.EnumType {
		writeEnum(w, edp, nil)
	}

	// Messages, recurse.
	for _, dp := range fdp.MessageType {
		writeDescriptor(w, dp, ns, nil)
	}

	// TODO: top level fields?

	// Write services.
	if genService {
		for _, sdp := range fdp.Service {
			writeService(w, sdp, fdp.GetPackage(), ns)
		}
	}

	// Write file descriptor.
	w.ln()
	fdClassName := strings.Replace(fdp.GetName(), "/", "_", -1)
	fdClassName = strings.Replace(fdClassName, ".", "__", -1)
	fdClassName = "__FileDescriptor_" + fdClassName
	w.p("class %s implements %s\\FileDescriptor {", fdClassName, libNsInternal)
	w.p("const string NAME = '%s';", fdp.GetName())
	w.p("const string RAW = '%s';", toPhpString(fdp))
	w.p("public function Name(): string {")
	w.p("return %s::NAME;", fdClassName)
	w.p("}")
	w.ln()
	w.p("public function FileDescriptorProtoBytes(): string {")
	w.p("return (string)gzuncompress(base64_decode(%s::RAW));", fdClassName)
	w.p("}")
	w.p("}")
}

func toPhpString(fdp *desc.FileDescriptorProto) string {
	bfdp, err := proto.Marshal(fdp)
	if err != nil {
		panic(err)
	}
	var b bytes.Buffer
	gz, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		panic(err)
	}
	if _, err = gz.Write(bfdp); err != nil {
		panic(err)
	}
	if err = gz.Flush(); err != nil {
		panic(err)
	}
	str := base64.RawStdEncoding.EncodeToString(b.Bytes())
	return str
}

func toPhpName(ns, name string) (string, string) {
	return strings.Replace(ns, ".", "\\", -1), strings.Replace(name, ".", "_", -1)
}

type field struct {
	fd                     *desc.FieldDescriptorProto
	typePhpNs, typePhpName string
	typeDescriptor         interface{}
	typeNs                 *Namespace
	typeEnumDefault        string
	isMap                  bool
}

func newField(fd *desc.FieldDescriptorProto, ns *Namespace) field {
	f := field{
		fd: fd,
	}
	if fd.GetTypeName() != "" {
		typeNs, typeName, i := ns.FindFullyQualifiedName(fd.GetTypeName())
		f.typePhpNs, f.typePhpName = toPhpName(typeNs, typeName)
		f.typeDescriptor = i
		f.typeNs = ns.FindFullyQualifiedNamespace(typeNs)
		if dp, ok := f.typeDescriptor.(*desc.DescriptorProto); ok {
			if dp.GetOptions().GetMapEntry() {
				f.isMap = true
			}
		}
		if ed, ok := f.typeDescriptor.(*desc.EnumDescriptorProto); ok {
			for _, v := range ed.Value {
				if v.GetNumber() == 0 {
					f.typeEnumDefault = v.GetName()
					break
				}
			}
		}

	}

	return f
}

func (f field) mapPhpTypes() (string, string) {
	dp := f.typeDescriptor.(*desc.DescriptorProto)
	keyField := newField(dp.Field[0], f.typeNs)
	valueField := newField(dp.Field[1], f.typeNs)
	return keyField.phpType(), valueField.labeledType()
}

func (f field) phpType() string {
	switch t := *f.fd.Type; t {
	case desc.FieldDescriptorProto_TYPE_STRING, desc.FieldDescriptorProto_TYPE_BYTES:
		return "string"
	case desc.FieldDescriptorProto_TYPE_INT64,
		desc.FieldDescriptorProto_TYPE_INT32, desc.FieldDescriptorProto_TYPE_UINT64, desc.FieldDescriptorProto_TYPE_UINT32, desc.FieldDescriptorProto_TYPE_SINT64, desc.FieldDescriptorProto_TYPE_SINT32, desc.FieldDescriptorProto_TYPE_FIXED32, desc.FieldDescriptorProto_TYPE_FIXED64, desc.FieldDescriptorProto_TYPE_SFIXED32, desc.FieldDescriptorProto_TYPE_SFIXED64:
		return "int"
	case desc.FieldDescriptorProto_TYPE_FLOAT, desc.FieldDescriptorProto_TYPE_DOUBLE:
		return "float"
	case desc.FieldDescriptorProto_TYPE_BOOL:
		return "bool"
	case desc.FieldDescriptorProto_TYPE_MESSAGE:
		return f.typePhpNs + "\\" + f.typePhpName
	case desc.FieldDescriptorProto_TYPE_ENUM:
		return f.typePhpNs + "\\" + f.typePhpName + "_EnumType"
	default:
		panic(fmt.Errorf("unexpected proto type while converting to php type: %v", t))
	}
}

func (f field) defaultValue() string {
	if f.isMap {
		return "dict[]"
	}
	if f.isRepeated() {
		return "vec[]"
	}
	switch t := *f.fd.Type; t {
	case desc.FieldDescriptorProto_TYPE_STRING, desc.FieldDescriptorProto_TYPE_BYTES:
		return "''"
	case desc.FieldDescriptorProto_TYPE_INT64,
		desc.FieldDescriptorProto_TYPE_INT32, desc.FieldDescriptorProto_TYPE_UINT64, desc.FieldDescriptorProto_TYPE_UINT32, desc.FieldDescriptorProto_TYPE_SINT64, desc.FieldDescriptorProto_TYPE_SINT32, desc.FieldDescriptorProto_TYPE_FIXED32, desc.FieldDescriptorProto_TYPE_FIXED64, desc.FieldDescriptorProto_TYPE_SFIXED32, desc.FieldDescriptorProto_TYPE_SFIXED64:
		return "0"
	case desc.FieldDescriptorProto_TYPE_FLOAT, desc.FieldDescriptorProto_TYPE_DOUBLE:
		return "0.0"
	case desc.FieldDescriptorProto_TYPE_BOOL:
		return "false"
	case desc.FieldDescriptorProto_TYPE_ENUM:
		return f.typePhpNs + "\\" + f.typePhpName + "::" + f.typeEnumDefault
	case desc.FieldDescriptorProto_TYPE_MESSAGE:
		return "null"
	default:
		panic(fmt.Errorf("unexpected proto type while converting to default value: %v", t))
	}
}

func (f field) isRepeated() bool {
	return *f.fd.Label == desc.FieldDescriptorProto_LABEL_REPEATED
}

func (f field) labeledType() string {
	if f.isMap {
		k, v := f.mapPhpTypes()
		return fmt.Sprintf("dict<%s, %s>", k, v)
	}
	if f.isRepeated() {
		return "vec<" + f.phpType() + ">"
	}
	if *f.fd.Type == desc.FieldDescriptorProto_TYPE_MESSAGE {
		return "?" + f.phpType()
	}
	return f.phpType()
}

func (f field) varName() string {
	return *f.fd.Name
}

// Default is 0
var writeWireType = map[desc.FieldDescriptorProto_Type]int{
	desc.FieldDescriptorProto_TYPE_FLOAT:    5,
	desc.FieldDescriptorProto_TYPE_DOUBLE:   1,
	desc.FieldDescriptorProto_TYPE_FIXED32:  5,
	desc.FieldDescriptorProto_TYPE_SFIXED32: 5,
	desc.FieldDescriptorProto_TYPE_FIXED64:  1,
	desc.FieldDescriptorProto_TYPE_SFIXED64: 1,
	desc.FieldDescriptorProto_TYPE_STRING:   2,
	desc.FieldDescriptorProto_TYPE_BYTES:    2,
	desc.FieldDescriptorProto_TYPE_MESSAGE:  2,
}

var isPackable = map[desc.FieldDescriptorProto_Type]bool{
	desc.FieldDescriptorProto_TYPE_INT64:    true,
	desc.FieldDescriptorProto_TYPE_INT32:    true,
	desc.FieldDescriptorProto_TYPE_UINT64:   true,
	desc.FieldDescriptorProto_TYPE_UINT32:   true,
	desc.FieldDescriptorProto_TYPE_SINT64:   true,
	desc.FieldDescriptorProto_TYPE_SINT32:   true,
	desc.FieldDescriptorProto_TYPE_FLOAT:    true,
	desc.FieldDescriptorProto_TYPE_DOUBLE:   true,
	desc.FieldDescriptorProto_TYPE_FIXED32:  true,
	desc.FieldDescriptorProto_TYPE_SFIXED32: true,
	desc.FieldDescriptorProto_TYPE_FIXED64:  true,
	desc.FieldDescriptorProto_TYPE_SFIXED64: true,
	desc.FieldDescriptorProto_TYPE_BOOL:     true,
	desc.FieldDescriptorProto_TYPE_ENUM:     true,
}

func (f field) writeDecoder(w *writer, dec, wt string) {
	if f.isMap {
		w.p("$obj = new %s();", f.phpType())
		w.p("$obj->MergeFrom(%s->readDecoder());", dec)
		w.p("$this->%s[$obj->key] = $obj->value;", f.varName())
		return
	}
	if *f.fd.Type == desc.FieldDescriptorProto_TYPE_MESSAGE {
		// This is different enough we handle it on it's own.
		if f.isRepeated() {
			w.p("$obj = new %s();", f.phpType())
			w.p("$obj->MergeFrom(%s->readDecoder());", dec)
			w.p("$this->%s []= $obj;", f.varName())
		} else {
			w.p("if ($this->%s == null) {", f.varName())
			w.p("$this->%s = new %s();", f.varName(), f.phpType())
			w.p("}")
			w.p("$this->%s->MergeFrom(%s->readDecoder());", f.varName(), dec)
		}
		return
	}

	// TODO should we do wiretype checking here?
	reader := ""
	switch *f.fd.Type {
	case desc.FieldDescriptorProto_TYPE_STRING, desc.FieldDescriptorProto_TYPE_BYTES:
		reader = fmt.Sprintf("%s->readString()", dec)
	case desc.FieldDescriptorProto_TYPE_INT64, desc.FieldDescriptorProto_TYPE_INT32, desc.FieldDescriptorProto_TYPE_UINT64, desc.FieldDescriptorProto_TYPE_UINT32:
		reader = fmt.Sprintf("%s->readVarInt128()", dec)
	case desc.FieldDescriptorProto_TYPE_SINT64, desc.FieldDescriptorProto_TYPE_SINT32:
		reader = fmt.Sprintf("%s->readVarInt128ZigZag()", dec)
	case desc.FieldDescriptorProto_TYPE_FLOAT:
		reader = fmt.Sprintf("%s->readFloat()", dec)
	case desc.FieldDescriptorProto_TYPE_DOUBLE:
		reader = fmt.Sprintf("%s->readDouble()", dec)
	case desc.FieldDescriptorProto_TYPE_FIXED32, desc.FieldDescriptorProto_TYPE_SFIXED32:
		reader = fmt.Sprintf("%s->readLittleEndianInt32()", dec)
	case desc.FieldDescriptorProto_TYPE_FIXED64, desc.FieldDescriptorProto_TYPE_SFIXED64:
		reader = fmt.Sprintf("%s->readLittleEndianInt64()", dec)
	case desc.FieldDescriptorProto_TYPE_BOOL:
		reader = fmt.Sprintf("%s->readBool()", dec)
	case desc.FieldDescriptorProto_TYPE_ENUM:

		reader = fmt.Sprintf("%s\\%s::FromInt(%s->readVarInt128())", f.typePhpNs, f.typePhpName, dec)
	default:
		panic(fmt.Errorf("unknown reader for fd type: %s", *f.fd.Type))
	}
	if !f.isRepeated() {
		w.p("$this->%s = %s;", f.varName(), reader)
		return
	}
	// Repeated
	packable := isPackable[*f.fd.Type]
	if packable {
		w.p("if (%s == 2) {", wt)
		w.p("$packed = %s->readDecoder();", dec)
		w.p("while (!$packed->isEOF()) {")
		w.pdebug("reading packed field")
		packedReader := strings.Replace(reader, dec, "$packed", 1) // Heh, kinda hacky.
		w.p("$this->%s []= %s;", f.varName(), packedReader)
		w.p("}")
		w.p("} else {")
	}
	w.p("$this->%s []= %s;", f.varName(), reader)
	if packable {
		w.p("}")
	}
}

func (f field) writeEncoder(w *writer, enc string) {
	if f.isMap {
		w.p("foreach ($this->%s as $k => $v) {", f.varName())
		w.p("$obj = new %s();", f.phpType())
		w.p("$obj->key = $k;")
		w.p("$obj->value = $v;")
		w.p("$nested = new %s\\Encoder();", libNsInternal)
		w.p("$obj->WriteTo($nested);")
		w.p("%s->writeEncoder($nested, %d);", enc, *f.fd.Number)
		w.p("}")
		return
	}

	if *f.fd.Type == desc.FieldDescriptorProto_TYPE_MESSAGE {
		// This is different enough we handle it on it's own.
		// TODO we could optimize to not to string copies.
		if f.isRepeated() {
			w.p("foreach ($this->%s as $msg) {", f.varName())
		} else {
			w.p("$msg = $this->%s;", f.varName())
			w.p("if ($msg != null) {")
		}
		w.p("$nested = new %s\\Encoder();", libNsInternal)
		w.p("$msg->WriteTo($nested);")
		w.p("%s->writeEncoder($nested, %d);", enc, *f.fd.Number)
		w.p("}")
		return
	}

	writer := ""
	switch *f.fd.Type {
	case desc.FieldDescriptorProto_TYPE_STRING, desc.FieldDescriptorProto_TYPE_BYTES:
		writer = fmt.Sprintf("%s->writeString($this->%s)", enc, f.varName())
	case desc.FieldDescriptorProto_TYPE_INT64, desc.FieldDescriptorProto_TYPE_INT32, desc.FieldDescriptorProto_TYPE_UINT64, desc.FieldDescriptorProto_TYPE_UINT32:
		writer = fmt.Sprintf("%s->writeVarInt128($this->%s)", enc, f.varName())
	case desc.FieldDescriptorProto_TYPE_SINT64, desc.FieldDescriptorProto_TYPE_SINT32:
		writer = fmt.Sprintf("%s->writeVarInt128ZigZag($this->%s)", enc, f.varName())
	case desc.FieldDescriptorProto_TYPE_FLOAT:
		writer = fmt.Sprintf("%s->writeFloat($this->%s)", enc, f.varName())
	case desc.FieldDescriptorProto_TYPE_DOUBLE:
		writer = fmt.Sprintf("%s->writeDouble($this->%s)", enc, f.varName())
	case desc.FieldDescriptorProto_TYPE_FIXED32, desc.FieldDescriptorProto_TYPE_SFIXED32:
		writer = fmt.Sprintf("%s->writeLittleEndianInt32($this->%s)", enc, f.varName())
	case desc.FieldDescriptorProto_TYPE_FIXED64, desc.FieldDescriptorProto_TYPE_SFIXED64:
		writer = fmt.Sprintf("%s->writeLittleEndianInt64($this->%s)", enc, f.varName())
	case desc.FieldDescriptorProto_TYPE_BOOL:
		writer = fmt.Sprintf("%s->writeBool($this->%s)", enc, f.varName())
	case desc.FieldDescriptorProto_TYPE_ENUM:
		writer = fmt.Sprintf("%s->writeVarInt128($this->%s)", enc, f.varName())
	default:
		panic(fmt.Errorf("unknown reader for fd type: %s", *f.fd.Type))
	}
	tagWriter := fmt.Sprintf("%s->writeTag(%d, %d);", enc, *f.fd.Number, writeWireType[*f.fd.Type])

	if !f.isRepeated() {
		w.p("if ($this->%s !== %s) {", f.varName(), f.defaultValue())
		w.p(tagWriter)
		w.p("%s;", writer)
		w.p("}")
		return
	}
	// Repeated
	// Heh, kinda hacky.
	repeatWriter := strings.Replace(writer, "$this->"+f.varName(), "$elem", 1)
	if isPackable[*f.fd.Type] {
		// Heh, kinda hacky.
		packedWriter := strings.Replace(repeatWriter, enc, "$packed", 1)
		w.p("$packed = new %s\\Encoder();", libNsInternal)
		w.p("foreach ($this->%s as $elem) {", f.varName())
		w.pdebug("writing packed")
		w.p("%s;", packedWriter)
		w.p("}")
		w.p("%s->writeEncoder($packed, %d);", enc, *f.fd.Number)
	} else {
		w.p("foreach ($this->%s as $elem) {", f.varName())
		w.p(tagWriter)
		w.p("%s;", repeatWriter)
		w.p("}")
	}
}

// writeEnum writes an enumeration type and constants definitions.
func writeEnum(w *writer, ed *desc.EnumDescriptorProto, prefixNames []string) {
	name := strings.Join(append(prefixNames, *ed.Name), "_")
	typename := name + "_EnumType"
	w.p("newtype %s as int = int;", typename)
	w.p("class %s {", name)
	for _, v := range ed.Value {
		w.p("const %s %s = %d;", typename, *v.Name, *v.Number)
	}
	w.p("public static function FromInt(int $i): %s {", typename)
	w.p("return $i;")
	w.p("}")
	w.p("}")
	w.ln()
}

type oneof struct {
	descriptor *desc.OneofDescriptorProto
	className  string
	typeName   string
	fields     []field
}

// writeOneofEnum writes an enumeration type and constants definitions for
// proto "oneof" annotations.
func writeOneofEnum(w *writer, oo *oneof) {
	w.p("newtype %s = int;", oo.typeName)
	w.p("class %s {", oo.className)
	w.p("const %s NONE = 0;", oo.typeName)
	for _, field := range oo.fields {
		w.p("const %s %s = %d;", oo.typeName, field.fd.GetName(), field.fd.GetNumber())
	}
	w.p("}")
	w.ln()
}

// https://github.com/golang/protobuf/blob/master/protoc-gen-go/descriptor/descriptor.pb.go
func writeDescriptor(w *writer, dp *desc.DescriptorProto, ns *Namespace, prefixNames []string) {
	nextNames := append(prefixNames, dp.GetName())
	name := strings.Join(nextNames, "_")

	// Nested Enums.
	for _, edp := range dp.EnumType {
		writeEnum(w, edp, nextNames)
	}

	// Wrap fields in our own struct.
	fields := []field{}
	for _, fd := range dp.Field {
		fields = append(fields, newField(fd, ns))
	}

	// Oneofs: first group each field by it's corresponding oneof.
	oneofFields := map[int32][]field{}
	for _, field := range fields {
		if field.fd.OneofIndex == nil {
			continue
		}
		i := field.fd.GetOneofIndex()
		l := oneofFields[i]
		l = append(l, field)
		oneofFields[i] = l
	}

	// Write a oneof enum.
	oneofs := []*oneof{}
	for i, od := range dp.OneofDecl {
		name := strings.Join(append(nextNames, *od.Name), "_")
		oo := &oneof{
			descriptor: od,
			className:  name,
			typeName:   name + "_OneofType",
			fields:     oneofFields[int32(i)],
		}
		oneofs = append(oneofs, oo)
		writeOneofEnum(w, oo)
	}

	// Nested Types.
	for _, ndp := range dp.NestedType {
		writeDescriptor(w, ndp, ns, nextNames)
	}

	w.p("// message %s", dp.GetName())
	w.p("class %s implements %s\\Message {", name, libNs)

	// Members
	for _, f := range fields {
		w.p("// field %s = %d", f.fd.GetName(), f.fd.GetNumber())
		w.p("public %s $%s;", f.labeledType(), f.varName())
	}
	w.ln()

	// Constructor.
	w.p("public function __construct() {")
	for _, f := range fields {
		w.p("$this->%s = %s;", f.varName(), f.defaultValue())
	}
	w.p("}")
	w.ln()

	// Now sort the fields by number.
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].fd.GetNumber() < fields[j].fd.GetNumber()
	})

	// MergeFrom function
	w.p("public function MergeFrom(%s\\Decoder $d): void {", libNsInternal)
	w.p("while (!$d->isEOF()){")
	w.p("list($fn, $wt) = $d->readTag();")
	w.p("switch ($fn) {")
	for _, f := range fields {
		w.p("case %d:", f.fd.GetNumber())
		w.i++
		w.pdebug("reading field:%d (%s) wiretype:$wt of %s", f.fd.GetNumber(), f.varName(), dp.GetName())
		f.writeDecoder(w, "$d", "$wt")
		w.pdebug("read field:%d (%s) of %s", f.fd.GetNumber(), f.varName(), dp.GetName())
		w.p("break;")
		w.i--
	}
	w.p("default:")
	w.i++
	w.pdebug("skipping unknown field:$fn wiretype:$wt")
	w.p("$d->skipWireType($wt);")
	w.i--
	w.p("}") // switch
	w.p("}") // while
	w.p("}") // function MergeFrom
	w.ln()

	// WriteTo function
	w.p("public function WriteTo(%s\\Encoder $e): void {", libNsInternal)
	for _, f := range fields {
		w.pdebug("maybe writing field:%d (%s) of %s", f.fd.GetNumber(), f.varName(), dp.GetName())
		f.writeEncoder(w, "$e")
		w.pdebug("maybe wrote field:%d (%s) of %s", f.fd.GetNumber(), f.varName(), dp.GetName())
	}
	w.p("}") // WriteToFunction

	// Oneof enum helpers.
	for _, oneof := range oneofs {
		w.ln()
		w.p("public function oneof_%s(): %s {", oneof.descriptor.GetName(), oneof.typeName)
		for _, field := range oneof.fields {
			w.p("if ($this->%s != %s) {", field.varName(), field.defaultValue())
			w.p("return %s::%s;", oneof.className, field.fd.GetName())
			w.p("}")
		}
		w.p("return %s::NONE;", oneof.className)
		w.p("}")
	}

	w.p("}") // class
	w.ln()
}

type method struct {
	mdp                                  *desc.MethodDescriptorProto
	PhpName, InputPhpName, OutputPhpName string
}

func newMethod(mdp *desc.MethodDescriptorProto, ns *Namespace) method {
	m := method{mdp: mdp}
	m.PhpName = mdp.GetName()
	tns, tn, _ := ns.FindFullyQualifiedName(mdp.GetInputType())
	tns, tn = toPhpName(tns, tn)
	m.InputPhpName = tns + "\\" + tn
	tns, tn, _ = ns.FindFullyQualifiedName(mdp.GetOutputType())
	tns, tn = toPhpName(tns, tn)
	m.OutputPhpName = tns + "\\" + tn
	return m
}

func (m method) isStreaming() bool {
	return m.mdp.GetClientStreaming() || m.mdp.GetServerStreaming()
}

func writeService(w *writer, sdp *desc.ServiceDescriptorProto, pkg string, ns *Namespace) {
	methods := []method{}
	for _, mdp := range sdp.Method {
		methods = append(methods, newMethod(mdp, ns))
	}
	fqname := sdp.GetName()
	if pkg != "" {
		fqname = pkg + "." + fqname
	}

	// Client
	w.p("class %sClient {", sdp.GetName())
	w.p("public function __construct(private \\Grpc\\ClientConn $cc) {")
	w.p("}")
	for _, m := range methods {
		if m.isStreaming() {
			continue
		}
		w.ln()
		w.p("public async function %s(\\Grpc\\Context $ctx, %s $in, \\Grpc\\CallOption ...$co): Awaitable<%s> {", m.PhpName, m.InputPhpName, m.OutputPhpName)
		w.p("$out = new %s();", m.OutputPhpName)
		w.p("await $this->cc->Invoke($ctx, '/%s/%s', $in, $out, ...$co);", fqname, m.mdp.GetName())
		w.p("return $out;")
		w.p("}")
	}
	w.p("}")
	w.ln()

	// Server
	w.p("interface %sServer {", sdp.GetName())
	for _, m := range methods {
		if m.isStreaming() {
			continue
		}
		w.p("public function %s(\\Grpc\\Context $ctx, %s $in): %s;", m.PhpName, m.InputPhpName, m.OutputPhpName)
	}
	w.p("}")
	w.ln()

	w.p("function Register%sServer(\\Grpc\\Server $server, %sServer $service): void {", sdp.GetName(), sdp.GetName())
	w.p("$methods = vec[];")
	for _, m := range methods {
		if m.isStreaming() {
			continue
		}
		w.p("$handler = function(\\Grpc\\Context $ctx, \\Grpc\\DecoderFunc $df): %s\\Message use ($service) {", libNs)
		w.p("$in = new %s();", m.InputPhpName)
		w.p("$df($in);")
		w.p("return $service->%s($ctx, $in);", m.PhpName)
		w.p("};")
		w.p("$methods []= new \\Grpc\\MethodDesc('%s', $handler);", m.PhpName)
	}
	w.p("$server->RegisterService(new \\Grpc\\ServiceDesc('%s', $methods));", sdp.GetName())
	w.p("}")
}

// writer is a little helper for output printing. It indents code
// appropriately among other things.
type writer struct {
	w io.Writer
	i int
}

func (w *writer) p(format string, a ...interface{}) {
	if strings.HasPrefix(format, "}") {
		w.i--
	}
	indent := strings.Repeat("  ", w.i)
	fmt.Fprintf(w.w, indent+format, a...)
	w.ln()
	if strings.HasSuffix(format, "{") {
		w.i++
	}
}

func (w *writer) ln() {
	fmt.Fprintln(w.w)
}

func (w *writer) pdebug(format string, a ...interface{}) {
	if !genDebug {
		return
	}
	w.p(fmt.Sprintf(`echo "DEBUG: %s\n";`, format), a...)
}
