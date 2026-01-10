import 'package:flutter/material.dart';

import '../../domain/entities/topic_source.dart';

class TopicSourceDropdown extends StatelessWidget {
  const TopicSourceDropdown({
    super.key,
    required this.value,
    required this.onChanged,
  });

  final TopicSource? value;
  final ValueChanged<TopicSource?> onChanged;

  @override
  Widget build(BuildContext context) {
    return DropdownButton<TopicSource?>(
      value: value,
      isExpanded: true,
      onChanged: onChanged,
      items: <DropdownMenuItem<TopicSource?>>[
        const DropdownMenuItem<TopicSource?>(
          value: null,
          child: Text('全部来源'),
        ),
        ...TopicSource.values.map(
          (s) => DropdownMenuItem<TopicSource?>(
            value: s,
            child: Text(s.label),
          ),
        ),
      ],
    );
  }
}
