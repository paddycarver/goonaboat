package com.goonaboat;

import android.annotation.SuppressLint;
import android.app.Fragment;
import android.os.Bundle;
import android.os.Handler;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.webkit.WebSettings;
import android.webkit.WebView;

public class PreviewFragment extends Fragment {
	@Override
	public View onCreateView(LayoutInflater inflater, ViewGroup container,
			Bundle savedInstanceState) {
		return inflater.inflate(R.layout.fragment_preview, container, false);
	}

	@SuppressLint("SetJavaScriptEnabled")
	@Override
	public void onActivityCreated(Bundle savedInstanceState) {
		super.onActivityCreated(savedInstanceState);
		loadURL("http://goonaboat.com", null);
	}

	public void setSlide(int indexh, int indexv, int fragment) {
		String url = "javascript:Reveal.slide(" + indexh + ", " + indexv;
		if (fragment < 0) {
			url += ", " + fragment;
		}
		url += ");";
		WebView preview = (WebView) getView();
		preview.loadUrl(url);
		adjust();
	}

	public void adjust() {
		WebView preview = (WebView) getView();
		preview.loadUrl("javascript:Reveal.next();");
	}

	public void setSlide(int indexh, int indexv) {
		this.setSlide(indexh, indexv, -1);
	}

	@SuppressLint("SetJavaScriptEnabled")
	protected void loadURL(String base, String session) {
		WebView preview = (WebView) getView();
		WebSettings settings = preview.getSettings();
		settings.setJavaScriptEnabled(true);
		String url = base;
		if (session != null) {
			url += session;
		}
		preview.loadUrl(url);
		Handler handler = new Handler();
		handler.postDelayed(new Runnable() {
			public void run() {
				adjust();
			}
		}, 200);
	}
}