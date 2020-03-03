<?hh // strict
namespace google\protobuf;

// Generated by the protocol buffer compiler.  DO NOT EDIT!
// Source: field_mask.proto

class FieldMask implements \Protobuf\Message {
  public vec<string> $paths;
  private string $XXX_unrecognized;

  public function __construct(shape(
    ?'paths' => vec<string>,
  ) $s = shape()) {
    $this->paths = $s['paths'] ?? vec[];
    $this->XXX_unrecognized = '';
  }

  public function MessageName(): string {
    return "google.protobuf.FieldMask";
  }

  public function MergeFrom(\Protobuf\Internal\Decoder $d): void {
    while (!$d->isEOF()){
      list($fn, $wt) = $d->readTag();
      switch ($fn) {
        case 1:
          $this->paths []= $d->readString();
          break;
        default:
          $d->skip($fn, $wt);
      }
    }
    $this->XXX_unrecognized = $d->skippedRaw();
  }

  public function WriteTo(\Protobuf\Internal\Encoder $e): void {
    foreach ($this->paths as $elem) {
      $e->writeTag(1, 2);
      $e->writeString($elem);
    }
    $e->writeRaw($this->XXX_unrecognized);
  }

  public function WriteJsonTo(\Protobuf\Internal\JsonEncoder $e): void {
    $e->writePrimitiveList('paths', 'paths', $this->paths);
  }

  public function MergeJsonFrom(mixed $m): void {
    if ($m === null) return;
    $d = \Protobuf\Internal\JsonDecoder::readObject($m);
    foreach ($d as $k => $v) {
      switch ($k) {
        case 'paths':
          foreach(\Protobuf\Internal\JsonDecoder::readList($v) as $vv) {
            $this->paths []= \Protobuf\Internal\JsonDecoder::readString($vv);
          }
          break;
        default:
        break;
      }
    }
  }

  public function CopyFrom(\Protobuf\Message $o): \Errors\Error {
    if (!($o is FieldMask)) {
      return \Errors\Errorf('CopyFrom failed: incorrect type received: %s', $o->MessageName());
    }
    $this->paths = $o->paths;
    $this->XXX_unrecognized = $o->XXX_unrecognized;
    return \Errors\Ok();
  }
}


class XXX_FileDescriptor_field_mask__proto implements \Protobuf\Internal\FileDescriptor {
  const string NAME = 'field_mask.proto';
  const string RAW =
  'eNriEkjLTM1Jic9NLM7WKyjKL8kX4k/Pz0/PSYXwkkrTlBS5ON1AinwTi7OFRLhYCxJLMo'
  .'olGBWYNTiDIBynTkYu4eT8XD00rU58cI0BIKEAxihLqJL0/JzEvHS9/KJ0/fTUPLAGfZg2'
  .'fYSbrBHMRUzM7gFOq5jk3CEmBEBV64Wn5uR45+WX54VUFqQWJ7GBjTEGBAAA//9ktkwp';
  public function Name(): string {
    return self::NAME;
  }

  public function FileDescriptorProtoBytes(): string {
    return (string)\gzuncompress(\base64_decode(self::RAW));
  }
}
