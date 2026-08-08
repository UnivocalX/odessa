import 'package:flutter_test/flutter_test.dart';

import 'package:frontend/app.dart';

void main() {
  testWidgets('Login page is shown by default', (WidgetTester tester) async {
    await tester.pumpWidget(const OdessaApp());

    expect(find.text('Odessa Login'), findsOneWidget);
    expect(find.text('Email'), findsOneWidget);
    expect(find.text('Password'), findsOneWidget);
  });
}
