class AppConstants {
  const AppConstants._();

  static const String defaultApiUrl = String.fromEnvironment(
    'ODESSA_API_BASE_URL',
    defaultValue: 'http://localhost:9090',
  );

  static const int minPasswordLength = 8;
  static const int blobSearchPageSize = 100;

  static const double cardPadding = 16;
  static const double sectionSpacing = 12;
}
