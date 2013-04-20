package com.goonaboat;

import java.io.IOException;

import android.accounts.Account;
import android.accounts.AccountManager;
import android.accounts.AccountManagerCallback;
import android.accounts.AccountManagerFuture;
import android.accounts.AuthenticatorException;
import android.accounts.OperationCanceledException;
import android.app.Activity;
import android.content.Intent;
import android.os.Bundle;
import android.os.Handler;
import android.os.Handler.Callback;
import android.os.Message;
import android.util.Log;
import android.view.Menu;
import android.view.MenuItem;
import android.widget.Toast;

import com.google.android.gms.auth.GoogleAuthUtil;
import com.google.android.gms.common.AccountPicker;

public class MainActivity extends Activity implements Callback,
		AccountManagerCallback<Bundle> {
	private static final int ACCOUNT_PICK_CODE = 1;
	private String session = null;
	private String token = null;

	@Override
	protected void onCreate(Bundle savedInstanceState) {
		super.onCreate(savedInstanceState);
		setContentView(R.layout.activity_main);
		if (token == null || session == null) {
			Intent intent = AccountPicker.newChooseAccountIntent(null, null,
					new String[] { GoogleAuthUtil.GOOGLE_ACCOUNT_TYPE }, false,
					null, null, null, null);
			startActivityForResult(intent, ACCOUNT_PICK_CODE);
		} else {
			PreviewFragment preview = (PreviewFragment) getFragmentManager()
					.findFragmentById(R.id.preview_fragment);
			preview.loadURL("http://goonaboat.com/", session);
			SlideFragment slides = (SlideFragment) getFragmentManager()
					.findFragmentById(R.id.slide_fragment);
			slides.loadURL("http://goonaboat.com/", session, token);
		}
	}

	@Override
	public boolean onCreateOptionsMenu(Menu menu) {
		// Inflate the menu; this adds items to the action bar if it is present.
		getMenuInflater().inflate(R.menu.main, menu);
		return true;
	}

	@Override
	public boolean onOptionsItemSelected(MenuItem item) {
		switch (item.getItemId()) {
		case R.id.action_account:
			Intent intent = AccountPicker.newChooseAccountIntent(null, null,
					new String[] { GoogleAuthUtil.GOOGLE_ACCOUNT_TYPE }, false,
					null, null, null, null);
			startActivityForResult(intent, ACCOUNT_PICK_CODE);
			return true;
		default:
			return super.onOptionsItemSelected(item);
		}
	}

	@Override
	protected void onActivityResult(int requestCode, int resultCode, Intent data) {
		super.onActivityResult(requestCode, resultCode, data);
		if (requestCode == ACCOUNT_PICK_CODE) {
			if (resultCode == RESULT_OK) {
				String email = data
						.getStringExtra(AccountManager.KEY_ACCOUNT_NAME);
				AccountManager am = AccountManager.get(this);
				Bundle options = new Bundle();
				Account[] accounts = am
						.getAccountsByType(GoogleAuthUtil.GOOGLE_ACCOUNT_TYPE);
				Account account = null;
				for (int i = 0; i < accounts.length; i++) {
					if (accounts[i].name.equals(email)) {
						account = accounts[i];
						break;
					}
				}
				am.getAuthToken(
						account,
						"oauth2:https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/userinfo.profile",
						options, this, this, new Handler(this));
				/*
				 * String token = null; try { token = GoogleAuthUtil .getToken(
				 * this, email,
				 * "https://www.googleapis.com/auth/userinfo.profile https://www.googleapis.com/auth/userinfo.email"
				 * ); } catch (GooglePlayServicesAvailabilityException playEx) {
				 * Dialog dialog = GooglePlayServicesUtil.getErrorDialog(
				 * playEx.getConnectionStatusCode(), this, ACCOUNT_REQ_CODE);
				 * dialog.show(); // Use the dialog to present to the user. }
				 * catch (UserRecoverableAuthException recoverableException) {
				 * Intent recoveryIntent = recoverableException.getIntent();
				 * startActivityForResult(recoveryIntent, ACCOUNT_REQ_CODE); //
				 * Use the intent in a custom dialog or just //
				 * startActivityForResult. } catch (GoogleAuthException authEx)
				 * { // This is likely unrecoverable. Log.e("Go On A Boat",
				 * "Unrecoverable authentication exception: " +
				 * authEx.getLocalizedMessage(), authEx); } catch (IOException
				 * ioEx) { Log.i("Go On A Boat", "transient error encountered: "
				 * + ioEx.getMessage()); } if (token != null) {
				 * Toast.makeText(this, token, Toast.LENGTH_LONG).show(); }
				 */
			}
		}
	}

	public void run(AccountManagerFuture<Bundle> result) {
		Bundle bundle;
		try {
			bundle = result.getResult();
			String token = bundle.getString(AccountManager.KEY_AUTHTOKEN);
			String name = bundle.getString(AccountManager.KEY_ACCOUNT_NAME);
			PreviewFragment preview = (PreviewFragment) getFragmentManager()
					.findFragmentById(R.id.preview_fragment);
			preview.loadURL("http://goonaboat.com/", name);
			SlideFragment slides = (SlideFragment) getFragmentManager()
					.findFragmentById(R.id.slide_fragment);
			slides.loadURL("http://goonaboat.com/", name, token);
		} catch (OperationCanceledException e) {
			Toast.makeText(this,
					"Please grant permission, or we can't log you in.",
					Toast.LENGTH_LONG).show();
		} catch (AuthenticatorException e) {
			Toast.makeText(this, e.getLocalizedMessage(), Toast.LENGTH_LONG)
					.show();
		} catch (IOException e) {
			Toast.makeText(this, e.getLocalizedMessage(), Toast.LENGTH_LONG)
					.show();
		}
	}

	public boolean handleMessage(Message msg) {

		// TODO Auto-generated method stub
		Toast.makeText(this, "Error obtaining auth token. Try again.",
				Toast.LENGTH_LONG).show();
		return false;
	}
	
	public void onSlideChanged(int indexh, int indexv) {
		Log.i("Go On A Boat", "Slide changed: "+String.valueOf(indexh)+ " " + String.valueOf(indexv));
		PreviewFragment preview = (PreviewFragment) getFragmentManager()
				.findFragmentById(R.id.preview_fragment);
		preview.setSlide(indexh, indexv);
	}
	
	public void onLoad() {
		Log.i("goonaboat", "Page loaded");
		PreviewFragment preview = (PreviewFragment) getFragmentManager()
				.findFragmentById(R.id.preview_fragment);
		preview.adjust();
	}
	
	public void onFragmentChanged(int indexh, int indexv, int fragment) {
		PreviewFragment preview = (PreviewFragment) getFragmentManager()
				.findFragmentById(R.id.preview_fragment);
		preview.setSlide(indexh, indexv, fragment);
	}
	
	public void onNotesChanged(String notes) {
		NotesFragment nf = (NotesFragment) getFragmentManager().findFragmentById(R.id.notes_fragment);
		nf.setNotes(notes);
	}
}