import 'package:flutter/material.dart';

import '../colors.dart';

class OdessaTheme {
  const OdessaTheme._();

  static ThemeData dark() {
    return ThemeData(
      useMaterial3: true,
      scaffoldBackgroundColor: OdessaColors.background,
      colorScheme: const ColorScheme.dark(
        primary: OdessaColors.primary,
        secondary: OdessaColors.surface,
        surface: OdessaColors.primary,
        onPrimary: Colors.white,
        onSecondary: Colors.white,
        onSurface: Colors.white,
      ),
      appBarTheme: const AppBarTheme(
        backgroundColor: OdessaColors.primary,
        foregroundColor: OdessaColors.accentDark,
        elevation: 0,
      ),
      cardTheme: CardThemeData(
        color: OdessaColors.primary,
        elevation: 0,
        margin: EdgeInsets.zero,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(18),
          side: BorderSide(
            color: OdessaColors.accentDark.withValues(alpha: 0.35),
          ),
        ),
      ),
      navigationRailTheme: NavigationRailThemeData(
        backgroundColor: OdessaColors.background,
        indicatorColor: OdessaColors.accentDark.withValues(alpha: 0.24),
        selectedIconTheme: const IconThemeData(color: Colors.white),
        unselectedIconTheme: IconThemeData(
          color: OdessaColors.accentDark.withValues(alpha: 0.78),
        ),
        selectedLabelTextStyle: const TextStyle(
          color: Colors.white,
          fontWeight: FontWeight.w700,
        ),
        unselectedLabelTextStyle: TextStyle(
          color: OdessaColors.accentDark.withValues(alpha: 0.82),
        ),
      ),
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          foregroundColor: Colors.white,
          side: const BorderSide(color: Colors.white54),
          textStyle: const TextStyle(fontWeight: FontWeight.w600),
          disabledForegroundColor: Colors.white70,
        ),
      ),
      filledButtonTheme: FilledButtonThemeData(
        style: FilledButton.styleFrom(
          foregroundColor: Colors.white,
          textStyle: const TextStyle(fontWeight: FontWeight.w700),
          disabledForegroundColor: Colors.white70,
          disabledBackgroundColor: Colors.white24,
        ),
      ),
      textButtonTheme: TextButtonThemeData(
        style: TextButton.styleFrom(
          foregroundColor: OdessaColors.accentDark,
          disabledForegroundColor: OdessaColors.accentDark.withValues(
            alpha: 0.6,
          ),
        ),
      ),
      iconButtonTheme: IconButtonThemeData(
        style: IconButton.styleFrom(
          foregroundColor: Colors.white,
          disabledForegroundColor: Colors.white70,
        ),
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: OdessaColors.background.withValues(alpha: 0.35),
        labelStyle: TextStyle(
          color: OdessaColors.accentDark.withValues(alpha: 0.95),
        ),
        hintStyle: TextStyle(
          color: OdessaColors.accentDark.withValues(alpha: 0.72),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide(
            color: OdessaColors.accentDark.withValues(alpha: 0.45),
          ),
        ),
        focusedBorder: const OutlineInputBorder(
          borderRadius: BorderRadius.all(Radius.circular(12)),
          borderSide: BorderSide(color: Colors.white, width: 1.6),
        ),
        errorBorder: const OutlineInputBorder(
          borderRadius: BorderRadius.all(Radius.circular(12)),
          borderSide: BorderSide(color: Colors.redAccent),
        ),
        focusedErrorBorder: const OutlineInputBorder(
          borderRadius: BorderRadius.all(Radius.circular(12)),
          borderSide: BorderSide(color: Colors.redAccent, width: 1.6),
        ),
      ),
      textSelectionTheme: TextSelectionThemeData(
        cursorColor: Colors.white,
        selectionColor: OdessaColors.accentDark.withValues(alpha: 0.55),
        selectionHandleColor: OdessaColors.accentDark,
      ),
    );
  }
}
