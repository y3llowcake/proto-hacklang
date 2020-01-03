<?hh // strict

// Generated by the protocol buffer compiler.  DO NOT EDIT!
// Source: example3.proto

class Donkey implements \Protobuf\Message {
  public string $hi;
  private string $XXX_unrecognized;

  public function __construct(shape(
    ?'hi' => string,
  ) $s = shape()) {
    $this->hi = $s['hi'] ?? '';
    $this->XXX_unrecognized = '';
  }

  public function MessageName(): string {
    return "Donkey";
  }

  public function MergeFrom(\Protobuf\Internal\Decoder $d): void {
    while (!$d->isEOF()){
      list($fn, $wt) = $d->readTag();
      switch ($fn) {
        case 1:
          $this->hi = $d->readString();
          break;
        default:
          $d->skip($fn, $wt);
      }
    }
    $this->XXX_unrecognized = $d->skippedRaw();
  }

  public function WriteTo(\Protobuf\Internal\Encoder $e): void {
    if ($this->hi !== '') {
      $e->writeTag(1, 2);
      $e->writeString($this->hi);
    }
    $e->writeRaw($this->XXX_unrecognized);
  }

  public function WriteJsonTo(\Protobuf\Internal\JsonEncoder $e): void {
    $e->writeString('hi', 'hi', $this->hi, false);
  }

  public function MergeJsonFrom(mixed $m): void {
    if ($m === null) return;
    $d = \Protobuf\Internal\JsonDecoder::readObject($m);
    foreach ($d as $k => $v) {
      switch ($k) {
        case 'hi':
          $this->hi = \Protobuf\Internal\JsonDecoder::readString($v);
          break;
      }
    }
  }

  public function CopyFrom(\Protobuf\Message $o): void {
    if (!($o is Donkey)) {
      throw new \Protobuf\ProtobufException('CopyFrom failed: incorrect type received');
    }
    $this->hi = $o->hi;
    $this->XXX_unrecognized = $o->XXX_unrecognized;
  }
}

class Funky_Monkey implements \Protobuf\Message {
  public string $hi;
  private string $XXX_unrecognized;

  public function __construct(shape(
    ?'hi' => string,
  ) $s = shape()) {
    $this->hi = $s['hi'] ?? '';
    $this->XXX_unrecognized = '';
  }

  public function MessageName(): string {
    return "Funky.Monkey";
  }

  public function MergeFrom(\Protobuf\Internal\Decoder $d): void {
    while (!$d->isEOF()){
      list($fn, $wt) = $d->readTag();
      switch ($fn) {
        case 1:
          $this->hi = $d->readString();
          break;
        default:
          $d->skip($fn, $wt);
      }
    }
    $this->XXX_unrecognized = $d->skippedRaw();
  }

  public function WriteTo(\Protobuf\Internal\Encoder $e): void {
    if ($this->hi !== '') {
      $e->writeTag(1, 2);
      $e->writeString($this->hi);
    }
    $e->writeRaw($this->XXX_unrecognized);
  }

  public function WriteJsonTo(\Protobuf\Internal\JsonEncoder $e): void {
    $e->writeString('hi', 'hi', $this->hi, false);
  }

  public function MergeJsonFrom(mixed $m): void {
    if ($m === null) return;
    $d = \Protobuf\Internal\JsonDecoder::readObject($m);
    foreach ($d as $k => $v) {
      switch ($k) {
        case 'hi':
          $this->hi = \Protobuf\Internal\JsonDecoder::readString($v);
          break;
      }
    }
  }

  public function CopyFrom(\Protobuf\Message $o): void {
    if (!($o is Funky_Monkey)) {
      throw new \Protobuf\ProtobufException('CopyFrom failed: incorrect type received');
    }
    $this->hi = $o->hi;
    $this->XXX_unrecognized = $o->XXX_unrecognized;
  }
}

class Funky implements \Protobuf\Message {
  public ?\Funky_Monkey $monkey;
  public ?\Donkey $dokey;
  private string $XXX_unrecognized;

  public function __construct(shape(
    ?'monkey' => ?\Funky_Monkey,
    ?'dokey' => ?\Donkey,
  ) $s = shape()) {
    $this->monkey = $s['monkey'] ?? null;
    $this->dokey = $s['dokey'] ?? null;
    $this->XXX_unrecognized = '';
  }

  public function MessageName(): string {
    return "Funky";
  }

  public function MergeFrom(\Protobuf\Internal\Decoder $d): void {
    while (!$d->isEOF()){
      list($fn, $wt) = $d->readTag();
      switch ($fn) {
        case 1:
          if ($this->monkey == null) $this->monkey = new \Funky_Monkey();
          $this->monkey->MergeFrom($d->readDecoder());
          break;
        case 2:
          if ($this->dokey == null) $this->dokey = new \Donkey();
          $this->dokey->MergeFrom($d->readDecoder());
          break;
        default:
          $d->skip($fn, $wt);
      }
    }
    $this->XXX_unrecognized = $d->skippedRaw();
  }

  public function WriteTo(\Protobuf\Internal\Encoder $e): void {
    $msg = $this->monkey;
    if ($msg != null) {
      $nested = new \Protobuf\Internal\Encoder();
      $msg->WriteTo($nested);
      $e->writeEncoder($nested, 1);
    }
    $msg = $this->dokey;
    if ($msg != null) {
      $nested = new \Protobuf\Internal\Encoder();
      $msg->WriteTo($nested);
      $e->writeEncoder($nested, 2);
    }
    $e->writeRaw($this->XXX_unrecognized);
  }

  public function WriteJsonTo(\Protobuf\Internal\JsonEncoder $e): void {
    $e->writeMessage('monkey', 'monkey', $this->monkey, false);
    $e->writeMessage('dokey', 'dokey', $this->dokey, false);
  }

  public function MergeJsonFrom(mixed $m): void {
    if ($m === null) return;
    $d = \Protobuf\Internal\JsonDecoder::readObject($m);
    foreach ($d as $k => $v) {
      switch ($k) {
        case 'monkey':
          if ($v === null) break;
          if ($this->monkey == null) $this->monkey = new \Funky_Monkey();
          $this->monkey->MergeJsonFrom($v);
          break;
        case 'dokey':
          if ($v === null) break;
          if ($this->dokey == null) $this->dokey = new \Donkey();
          $this->dokey->MergeJsonFrom($v);
          break;
      }
    }
  }

  public function CopyFrom(\Protobuf\Message $o): void {
    if (!($o is Funky)) {
      throw new \Protobuf\ProtobufException('CopyFrom failed: incorrect type received');
    }
    $tmp = $o->monkey;
    if ($tmp !== null) {
      $nv = new \Funky_Monkey();
      $nv->CopyFrom($tmp);
      $this->monkey = $nv;
    }
    $tmp = $o->dokey;
    if ($tmp !== null) {
      $nv = new \Donkey();
      $nv->CopyFrom($tmp);
      $this->dokey = $nv;
    }
    $this->XXX_unrecognized = $o->XXX_unrecognized;
  }
}


class XXX_FileDescriptor_example3__proto implements \Protobuf\Internal\FileDescriptor {
  const string NAME = 'example3.proto';
  const string RAW =
  'eNri4kutSMwtyEk11isoyi/JV5LgYnPJz8tOrRTi42LKyJRgVGDU4AxiyshUSudidSvNy6'
  .'4UUuViywUrAUtyG/HqgcX1fMGCQVBJIVku1pR8kComsCp2PYi5QRBRKQkuNl+sFiWxgV1i'
  .'DAgAAP//iA4o6g';
  public function Name(): string {
    return self::NAME;
  }

  public function FileDescriptorProtoBytes(): string {
    return (string)\gzuncompress(\base64_decode(self::RAW));
  }
}
