package com.goonaboat;

import android.annotation.SuppressLint;
import android.app.Fragment;
import android.os.Bundle;
import android.util.Log;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.webkit.JavascriptInterface;
import android.webkit.WebSettings;
import android.webkit.WebView;

public class SlideFragment extends Fragment {
	public SlideFragment() {
	}

	@Override
	public View onCreateView(LayoutInflater inflater, ViewGroup container,
			Bundle savedInstanceState) {
		return inflater.inflate(R.layout.fragment_slide, container, false);
	}

	@Override
	public void onActivityCreated(Bundle savedInstanceState) {
		super.onActivityCreated(savedInstanceState);
		loadURL("http://goonaboat.com/", null, null);
	}

	@SuppressLint("SetJavaScriptEnabled")
	protected void loadURL(String base, String session, String auth) {
		WebView slide = (WebView) getView();
		WebSettings settings = slide.getSettings();
		settings.setJavaScriptEnabled(true);
		settings.setUserAgentString(settings.getUserAgentString() + " "
				+ "/GoOnABoat/1.0");
		MainActivity activity = (MainActivity) getActivity();
		slide.addJavascriptInterface(new JsObject(activity), "Android");
		String url = base;
		if (session != null && auth != null) {
			url += session + "?auth=" + auth;
		}
		Log.i("goonaboat", url);
		slide.loadUrl(url);
	}

	class JsObject {
		MainActivity activity;

		protected JsObject(MainActivity activity) {
			this.activity = activity;
		}

		@JavascriptInterface
		public void onSlideChanged(int indexh, int indexv) {
			Log.i("GoOnABoat", "Slide changed!");
			activity.onSlideChanged(indexh, indexv);
		}
		
		@JavascriptInterface
		public void onNotesChanged(String notes) {
			activity.onNotesChanged(notes);
		}
		
		@JavascriptInterface
		public void onLoad() {
			Log.i("Go On A Boat", "page loaded");
			activity.onLoad();
		}

		@JavascriptInterface
		public void onFragmentChanged(int indexh, int indexv, int fragment) {
			activity.onFragmentChanged(indexh, indexv, fragment);
		}
	}
}
