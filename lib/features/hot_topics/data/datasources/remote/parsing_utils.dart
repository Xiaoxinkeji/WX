Uri? tryParseUri(Object? value) {
  if (value == null) return null;
  if (value is Uri) return value;
  if (value is String) return Uri.tryParse(value);
  return Uri.tryParse(value.toString());
}

String? asString(Object? value) {
  if (value == null) return null;
  if (value is String) return value;
  return value.toString();
}

int? asInt(Object? value) {
  if (value == null) return null;
  if (value is int) return value;
  if (value is num) return value.toInt();
  final text = value.toString().trim();
  return int.tryParse(text);
}

num? asNum(Object? value) {
  if (value == null) return null;
  if (value is num) return value;
  final text = value.toString().trim();
  return num.tryParse(text);
}

Map<String, Object?>? asMap(Object? value) {
  if (value is Map<String, Object?>) return value;
  if (value is Map) {
    final out = <String, Object?>{};
    value.forEach((k, v) {
      if (k is String) out[k] = v;
    });
    return out;
  }
  return null;
}

List<Object?>? asList(Object? value) {
  if (value is List<Object?>) return value;
  if (value is List) return value.cast<Object?>();
  return null;
}

num? tryParseNumFromText(String? text) {
  if (text == null) return null;
  final match = RegExp(r'(\d+(?:\.\d+)?)').firstMatch(text.replaceAll(',', ''));
  if (match == null) return null;
  return num.tryParse(match.group(1)!);
}

String normalizeQuery(String query) => query.trim().toLowerCase();

bool containsIgnoreCase(String text, String query) {
  return text.toLowerCase().contains(normalizeQuery(query));
}
