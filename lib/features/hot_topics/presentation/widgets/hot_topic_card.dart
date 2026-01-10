import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../domain/entities/hot_topic.dart';

class HotTopicCard extends StatelessWidget {
  const HotTopicCard({
    super.key,
    required this.topic,
  });

  final HotTopic topic;

  @override
  Widget build(BuildContext context) {
    final subtitleParts = <String>[];
    if (topic.hotValue != null) subtitleParts.add('热度 ${topic.hotValue}');
    if (topic.url != null) subtitleParts.add(topic.url.toString());

    return Card(
      child: ListTile(
        leading: CircleAvatar(child: Text('${topic.rank}')),
        title: Text(topic.title, maxLines: 2, overflow: TextOverflow.ellipsis),
        subtitle: subtitleParts.isEmpty
            ? null
            : Text(
                subtitleParts.join(' · '),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
        trailing: Text(topic.source.label),
        onTap: topic.url == null
            ? null
            : () async {
                await Clipboard.setData(ClipboardData(text: topic.url.toString()));
                if (!context.mounted) return;
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(content: Text('链接已复制到剪贴板')),
                );
              },
      ),
    );
  }
}
